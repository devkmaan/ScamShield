from __future__ import annotations

import json
import math
import os
import re
from collections import Counter, defaultdict
from dataclasses import dataclass
from pathlib import Path
from typing import Any
from urllib.parse import urlparse

MODEL_VERSION = "text-scam-nb-v1"
MINILM_MODEL_VERSION = "text-scam-minilm-logreg-v1"
URL_MODEL_VERSION = "url-lexical-v1"
CALIBRATION_VERSION = "holdout-bucket-calibration-v1"
MINILM_CALIBRATION_VERSION = "minilm-holdout-bucket-calibration-v1"
DEFAULT_EMBEDDING_MODEL = "sentence-transformers/paraphrase-multilingual-MiniLM-L12-v2"
TOKEN_RE = re.compile(r"[a-z0-9@._-]+")
SCAM_TYPES = [
    "UPI_COLLECT",
    "PHISHING",
    "IMPERSONATION",
    "JOB_SCAM",
    "INVESTMENT",
    "LOAN_APP",
    "FAKE_RECEIPT",
]


def tokenize(text: str) -> list[str]:
    normalized = text.lower().replace("₹", " inr ")
    tokens = TOKEN_RE.findall(normalized)
    bigrams = [tokens[i] + "_" + tokens[i + 1] for i in range(len(tokens) - 1)]
    return tokens + bigrams


def load_samples(path: Path) -> list[dict[str, str]]:
    samples: list[dict[str, str]] = []
    with path.open("r", encoding="utf-8") as fh:
        for line in fh:
            line = line.strip()
            if line:
                samples.append(json.loads(line))
    return samples


@dataclass
class Evaluation:
    accuracy: float
    precision: float
    recall: float
    f1: float
    ece: float
    support: int
    calibration: list[dict[str, float]]


def backend_requested() -> str:
    requested = os.getenv("SCAMSHIELD_ML_BACKEND", "auto").strip().lower()
    if requested in {"naive_bayes", "nb", "fallback"}:
        return "naive_bayes"
    if requested in {"minilm", "embeddings", "sentence_transformers"}:
        return "minilm"
    return "auto"


def require_minilm() -> bool:
    value = os.getenv("SCAMSHIELD_REQUIRE_MINILM", "").strip().lower()
    return value in {"1", "true", "yes", "y"}


def embedding_model_name() -> str:
    return os.getenv("SCAMSHIELD_EMBEDDING_MODEL", DEFAULT_EMBEDDING_MODEL).strip() or DEFAULT_EMBEDDING_MODEL


def build_calibration(scores: list[float], actuals: list[int]) -> tuple[list[dict[str, float]], float]:
    if not scores:
        return ([{"bucket": bucket, "avgPrediction": 0.0, "observedRate": 0.0, "support": 0} for bucket in range(5)], 0.0)
    buckets: dict[int, list[tuple[float, int]]] = defaultdict(list)
    for score, actual in zip(scores, actuals):
        bucket = min(4, int(score * 5))
        buckets[bucket].append((score, actual))
    calibration: list[dict[str, float]] = []
    ece = 0.0
    for bucket in range(5):
        rows = buckets.get(bucket, [])
        if not rows:
            calibration.append({"bucket": bucket, "avgPrediction": 0.0, "observedRate": 0.0, "support": 0})
            continue
        avg_pred = sum(row[0] for row in rows) / len(rows)
        observed = sum(row[1] for row in rows) / len(rows)
        ece += (len(rows) / len(scores)) * abs(avg_pred - observed)
        calibration.append({
            "bucket": bucket,
            "avgPrediction": round(avg_pred, 4),
            "observedRate": round(observed, 4),
            "support": len(rows),
        })
    return calibration, round(ece, 4)


def metrics_from_scores(scores: list[float], actuals: list[int], threshold: float = 0.35) -> tuple[float, float, float, float]:
    tp = fp = tn = fn = 0
    for score, actual in zip(scores, actuals):
        pred = 1 if score >= threshold else 0
        if pred == 1 and actual == 1:
            tp += 1
        elif pred == 1 and actual == 0:
            fp += 1
        elif pred == 0 and actual == 0:
            tn += 1
        else:
            fn += 1
    precision = tp / max(1, tp + fp)
    recall = tp / max(1, tp + fn)
    f1 = (2 * precision * recall) / max(0.0001, precision + recall)
    accuracy = (tp + tn) / max(1, len(scores))
    return round(accuracy, 4), round(precision, 4), round(recall, 4), round(f1, 4)


def text_signals(text: str, tokens: list[str], score: float, type_scores: dict[str, float]) -> list[str]:
    signals: list[str] = []
    token_set = set(tokens)
    if {"otp", "share"}.issubset(token_set) or "share_otp" in token_set:
        signals.append("ml_otp_sharing")
    if "kyc" in token_set or "account_blocked" in token_set:
        signals.append("ml_kyc_pressure")
    if "upi" in token_set and ("pin" in token_set or "upi_pin" in token_set):
        signals.append("ml_upi_pin_context")
    if "guaranteed" in token_set or "double_money" in token_set:
        signals.append("ml_unrealistic_returns")
    if "task" in token_set and ("deposit" in token_set or "commission" in token_set):
        signals.append("ml_task_deposit")
    if re.search(r"https?://|www\.", text, flags=re.I):
        signals.append("ml_url_present")
    if score >= 0.75:
        signals.append("ml_high_probability")
    top_type = max(type_scores.items(), key=lambda item: item[1])[0] if type_scores else "UNKNOWN"
    if top_type != "UNKNOWN":
        signals.append("ml_type_" + top_type.lower())
    return signals[:8]


def heuristic_probability(tokens: list[str]) -> float:
    token_set = set(tokens)
    score = 0.08
    weighted_patterns = [
        ({"otp", "share"}, 0.38),
        ({"kyc"}, 0.22),
        ({"blocked"}, 0.18),
        ({"upi", "pin"}, 0.42),
        ({"qr", "receive"}, 0.34),
        ({"refund", "receive"}, 0.28),
        ({"safe", "account"}, 0.40),
        ({"fraud", "alert"}, 0.32),
        ({"anydesk"}, 0.38),
        ({"task", "deposit"}, 0.34),
        ({"registration", "fee"}, 0.28),
        ({"guaranteed", "return"}, 0.34),
        ({"double", "money"}, 0.34),
        ({"crypto", "profit"}, 0.26),
        ({"loan", "overdue"}, 0.22),
        ({"contact", "list"}, 0.22),
    ]
    for required, weight in weighted_patterns:
        if required.issubset(token_set):
            score += weight * (1 - score)
    bigram_hits = {
        "share_otp",
        "upi_pin",
        "safe_account",
        "double_money",
        "guaranteed_return",
        "registration_fee",
        "account_blocked",
        "receive_money",
    }
    for hit in bigram_hits:
        if hit in token_set:
            score += 0.24 * (1 - score)
    return min(0.99, score)


class NaiveBayesScamModel:
    backend_name = "naive_bayes"
    model_version = MODEL_VERSION
    calibration_version = CALIBRATION_VERSION
    embedding_model = ""

    def __init__(self, fallback_reason: str = "") -> None:
        self.vocab: set[str] = set()
        self.binary_counts = {"scam": Counter(), "safe": Counter()}
        self.binary_docs = Counter()
        self.binary_tokens = Counter()
        self.type_counts: dict[str, Counter[str]] = {name: Counter() for name in SCAM_TYPES}
        self.type_docs = Counter()
        self.type_tokens = Counter()
        self.calibration_bins: list[dict[str, float]] = []
        self.evaluation = Evaluation(0, 0, 0, 0, 0, 0, [])
        self.fallback_reason = fallback_reason

    def train(self, samples: list[dict[str, str]]) -> None:
        holdout = [sample for idx, sample in enumerate(samples) if idx % 5 == 0]
        train = [sample for idx, sample in enumerate(samples) if idx % 5 != 0]
        if not holdout:
            holdout = train[:]
        for sample in train:
            label = sample["label"]
            scam_type = sample.get("scamType", "UNKNOWN")
            tokens = tokenize(sample["text"])
            self.binary_docs[label] += 1
            self.binary_counts[label].update(tokens)
            self.binary_tokens[label] += len(tokens)
            self.vocab.update(tokens)
            if label == "scam" and scam_type in self.type_counts:
                self.type_docs[scam_type] += 1
                self.type_counts[scam_type].update(tokens)
                self.type_tokens[scam_type] += len(tokens)
        self.evaluation = self.evaluate(holdout)
        self.calibration_bins = self.evaluation.calibration

    def predict_text(self, text: str) -> dict[str, Any]:
        tokens = tokenize(text)
        raw_score = self._binary_probability(tokens)
        score = self._calibrate(raw_score)
        type_scores = self._type_scores(tokens, score)
        signals = self._signals(text, tokens, score, type_scores)
        confidence = min(0.95, 0.52 + abs(score - 0.5) * 0.65 + min(len(tokens), 30) * 0.004)
        return {
            "modelVersion": self.model_version,
            "score": round(score, 4),
            "confidence": round(confidence, 4),
            "scamTypeScores": type_scores,
            "signals": signals,
            "calibrationVersion": self.calibration_version,
        }

    def evaluate(self, samples: list[dict[str, str]]) -> Evaluation:
        if not samples:
            return Evaluation(0, 0, 0, 0, 0, 0, [])
        tp = fp = tn = fn = 0
        buckets: dict[int, list[tuple[float, int]]] = defaultdict(list)
        for sample in samples:
            actual = 1 if sample["label"] == "scam" else 0
            score = self._binary_probability(tokenize(sample["text"]))
            pred = 1 if score >= 0.35 else 0
            if pred == 1 and actual == 1:
                tp += 1
            elif pred == 1 and actual == 0:
                fp += 1
            elif pred == 0 and actual == 0:
                tn += 1
            else:
                fn += 1
            bucket = min(4, int(score * 5))
            buckets[bucket].append((score, actual))
        precision = tp / max(1, tp + fp)
        recall = tp / max(1, tp + fn)
        f1 = (2 * precision * recall) / max(0.0001, precision + recall)
        accuracy = (tp + tn) / max(1, len(samples))
        calibration: list[dict[str, float]] = []
        ece = 0.0
        for bucket in range(5):
            rows = buckets.get(bucket, [])
            if not rows:
                calibration.append({"bucket": bucket, "avgPrediction": 0.0, "observedRate": 0.0, "support": 0})
                continue
            avg_pred = sum(row[0] for row in rows) / len(rows)
            observed = sum(row[1] for row in rows) / len(rows)
            ece += (len(rows) / len(samples)) * abs(avg_pred - observed)
            calibration.append({
                "bucket": bucket,
                "avgPrediction": round(avg_pred, 4),
                "observedRate": round(observed, 4),
                "support": len(rows),
            })
        return Evaluation(
            accuracy=round(accuracy, 4),
            precision=round(precision, 4),
            recall=round(recall, 4),
            f1=round(f1, 4),
            ece=round(ece, 4),
            support=len(samples),
            calibration=calibration,
        )

    def _binary_probability(self, tokens: list[str]) -> float:
        scam_log = self._class_log_prob("scam", tokens)
        safe_log = self._class_log_prob("safe", tokens)
        max_log = max(scam_log, safe_log)
        scam_exp = math.exp(scam_log - max_log)
        safe_exp = math.exp(safe_log - max_log)
        nb_score = scam_exp / (scam_exp + safe_exp)
        heuristic = heuristic_probability(tokens)
        return min(0.99, max(0.01, 0.58 * nb_score + 0.42 * heuristic))

    def _class_log_prob(self, label: str, tokens: list[str]) -> float:
        total_docs = self.binary_docs["scam"] + self.binary_docs["safe"]
        prior = (self.binary_docs[label] + 1) / (total_docs + 2)
        vocab_size = max(1, len(self.vocab))
        denom = self.binary_tokens[label] + vocab_size
        score = math.log(prior)
        for token in tokens:
            score += math.log((self.binary_counts[label][token] + 1) / denom)
        return score

    def _type_scores(self, tokens: list[str], scam_score: float) -> dict[str, float]:
        if scam_score < 0.35 or not tokens:
            return {"UNKNOWN": round(1 - scam_score, 4)}
        logs: dict[str, float] = {}
        total_docs = sum(self.type_docs.values())
        vocab_size = max(1, len(self.vocab))
        for scam_type in SCAM_TYPES:
            prior = (self.type_docs[scam_type] + 1) / (total_docs + len(SCAM_TYPES))
            denom = self.type_tokens[scam_type] + vocab_size
            value = math.log(prior)
            for token in tokens:
                value += math.log((self.type_counts[scam_type][token] + 1) / denom)
            logs[scam_type] = value
        max_log = max(logs.values())
        exps = {key: math.exp(value - max_log) for key, value in logs.items()}
        total = sum(exps.values())
        return {key: round((value / total) * scam_score, 4) for key, value in exps.items()}

    def _calibrate(self, raw_score: float) -> float:
        bucket = min(4, int(raw_score * 5))
        if bucket >= len(self.calibration_bins):
            return raw_score
        calibration = self.calibration_bins[bucket]
        if calibration.get("support", 0) < 2:
            return raw_score
        observed = calibration["observedRate"]
        return min(0.99, max(0.01, 0.72 * raw_score + 0.28 * observed))

    def _signals(self, text: str, tokens: list[str], score: float, type_scores: dict[str, float]) -> list[str]:
        return text_signals(text, tokens, score, type_scores)

    def _heuristic_probability(self, tokens: list[str]) -> float:
        return heuristic_probability(tokens)

    def metadata(self) -> dict[str, Any]:
        return {
            "backend": self.backend_name,
            "modelVersion": self.model_version,
            "calibrationVersion": self.calibration_version,
            "embeddingModel": self.embedding_model,
            "fallbackReason": self.fallback_reason,
        }


class MiniLMCalibratedScamModel:
    backend_name = "minilm_embeddings"
    model_version = MINILM_MODEL_VERSION
    calibration_version = MINILM_CALIBRATION_VERSION

    def __init__(self, model_name: str | None = None) -> None:
        self.embedding_model = model_name or embedding_model_name()
        self.encoder: Any = None
        self.binary_classifier: Any = None
        self.type_classifier: Any = None
        self.calibration_bins: list[dict[str, float]] = []
        self.evaluation = Evaluation(0, 0, 0, 0, 0, 0, [])
        self.fallback_reason = ""

    def train(self, samples: list[dict[str, str]]) -> None:
        try:
            from sentence_transformers import SentenceTransformer
            from sklearn.linear_model import LogisticRegression
        except Exception as exc:  # pragma: no cover - exercised on machines without optional deps.
            raise RuntimeError(f"MiniLM dependencies are unavailable: {exc}") from exc

        train = [sample for idx, sample in enumerate(samples) if idx % 5 != 0]
        holdout = [sample for idx, sample in enumerate(samples) if idx % 5 == 0]
        if not holdout:
            holdout = train[:]
        if len({sample["label"] for sample in train}) < 2:
            raise RuntimeError("MiniLM classifier needs at least two labels in the training data")

        self.encoder = SentenceTransformer(self.embedding_model)
        train_texts = [sample["text"] for sample in train]
        train_vectors = self._encode(train_texts)
        train_labels = [sample["label"] for sample in train]

        self.binary_classifier = LogisticRegression(max_iter=1000, class_weight="balanced", random_state=42)
        self.binary_classifier.fit(train_vectors, train_labels)

        scam_train = [sample for sample in train if sample["label"] == "scam" and sample.get("scamType") in SCAM_TYPES]
        if len({sample["scamType"] for sample in scam_train}) >= 2:
            self.type_classifier = LogisticRegression(max_iter=1000, class_weight="balanced", random_state=42)
            self.type_classifier.fit(
                self._encode([sample["text"] for sample in scam_train]),
                [sample["scamType"] for sample in scam_train],
            )

        holdout_scores = self._raw_scores([sample["text"] for sample in holdout])
        holdout_actuals = [1 if sample["label"] == "scam" else 0 for sample in holdout]
        self.calibration_bins, _ = build_calibration(holdout_scores, holdout_actuals)
        self.evaluation = self.evaluate(holdout)

    def predict_text(self, text: str) -> dict[str, Any]:
        clean_text = text.strip()
        if not clean_text:
            return {
                "modelVersion": self.model_version,
                "score": 0.05,
                "confidence": 0.45,
                "scamTypeScores": {"UNKNOWN": 0.95},
                "signals": ["embedding_minilm", "calibrated_classifier"],
                "calibrationVersion": self.calibration_version,
            }
        tokens = tokenize(clean_text)
        raw_score = self._raw_scores([clean_text])[0]
        heuristic = heuristic_probability(tokens)
        score = self._calibrate(min(0.99, max(0.01, 0.82 * raw_score + 0.18 * heuristic)))
        type_scores = self._type_scores(clean_text, score)
        signals = ["embedding_minilm", "calibrated_classifier"] + text_signals(clean_text, tokens, score, type_scores)
        confidence = min(0.96, 0.57 + abs(score - 0.5) * 0.62 + min(len(tokens), 28) * 0.003)
        return {
            "modelVersion": self.model_version,
            "score": round(score, 4),
            "confidence": round(confidence, 4),
            "scamTypeScores": type_scores,
            "signals": signals[:8],
            "calibrationVersion": self.calibration_version,
        }

    def evaluate(self, samples: list[dict[str, str]]) -> Evaluation:
        if not samples:
            return Evaluation(0, 0, 0, 0, 0, 0, [])
        raw_scores = self._raw_scores([sample["text"] for sample in samples])
        calibrated_scores = [self._calibrate(score) for score in raw_scores]
        actuals = [1 if sample["label"] == "scam" else 0 for sample in samples]
        accuracy, precision, recall, f1 = metrics_from_scores(calibrated_scores, actuals)
        calibration, ece = build_calibration(calibrated_scores, actuals)
        return Evaluation(
            accuracy=accuracy,
            precision=precision,
            recall=recall,
            f1=f1,
            ece=ece,
            support=len(samples),
            calibration=calibration,
        )

    def metadata(self) -> dict[str, Any]:
        return {
            "backend": self.backend_name,
            "modelVersion": self.model_version,
            "calibrationVersion": self.calibration_version,
            "embeddingModel": self.embedding_model,
            "fallbackReason": self.fallback_reason,
        }

    def _encode(self, texts: list[str]) -> Any:
        if self.encoder is None:
            raise RuntimeError("MiniLM encoder has not been initialized")
        return self.encoder.encode(
            texts,
            batch_size=16,
            convert_to_numpy=True,
            normalize_embeddings=True,
            show_progress_bar=False,
        )

    def _raw_scores(self, texts: list[str]) -> list[float]:
        if self.binary_classifier is None:
            raise RuntimeError("MiniLM binary classifier has not been trained")
        vectors = self._encode(texts)
        classes = list(self.binary_classifier.classes_)
        scam_index = classes.index("scam")
        probabilities = self.binary_classifier.predict_proba(vectors)
        return [float(row[scam_index]) for row in probabilities]

    def _type_scores(self, text: str, scam_score: float) -> dict[str, float]:
        if scam_score < 0.35:
            return {"UNKNOWN": round(1 - scam_score, 4)}
        if self.type_classifier is None:
            return self._heuristic_type_scores(tokenize(text), scam_score)
        vector = self._encode([text])
        probabilities = self.type_classifier.predict_proba(vector)[0]
        scores: dict[str, float] = {}
        for scam_type, probability in zip(self.type_classifier.classes_, probabilities):
            scores[str(scam_type)] = round(float(probability) * scam_score, 4)
        return scores or self._heuristic_type_scores(tokenize(text), scam_score)

    def _heuristic_type_scores(self, tokens: list[str], scam_score: float) -> dict[str, float]:
        token_set = set(tokens)
        weights = {
            "UPI_COLLECT": 0.1,
            "PHISHING": 0.1,
            "IMPERSONATION": 0.1,
            "JOB_SCAM": 0.1,
            "INVESTMENT": 0.1,
            "LOAN_APP": 0.1,
            "FAKE_RECEIPT": 0.1,
        }
        if {"upi", "pin"}.issubset(token_set) or {"qr", "receive"}.issubset(token_set):
            weights["UPI_COLLECT"] += 0.8
        if "kyc" in token_set or "otp" in token_set or "login" in token_set or "verify" in token_set:
            weights["PHISHING"] += 0.8
        if "bank" in token_set or "rbi" in token_set or "police" in token_set:
            weights["IMPERSONATION"] += 0.6
        if "task" in token_set or "commission" in token_set or "salary" in token_set:
            weights["JOB_SCAM"] += 0.7
        if "crypto" in token_set or "trading" in token_set or "profit" in token_set:
            weights["INVESTMENT"] += 0.7
        if "loan" in token_set or "overdue" in token_set:
            weights["LOAN_APP"] += 0.7
        if "receipt" in token_set or "screenshot" in token_set:
            weights["FAKE_RECEIPT"] += 0.5
        total = sum(weights.values())
        return {key: round((value / total) * scam_score, 4) for key, value in weights.items()}

    def _calibrate(self, raw_score: float) -> float:
        bucket = min(4, int(raw_score * 5))
        if bucket >= len(self.calibration_bins):
            return raw_score
        calibration = self.calibration_bins[bucket]
        if calibration.get("support", 0) < 2:
            return raw_score
        observed = calibration["observedRate"]
        return min(0.99, max(0.01, 0.74 * raw_score + 0.26 * observed))


def score_url(url: str, context_text: str = "") -> dict[str, Any]:
    candidate = url.strip()
    if candidate and not re.match(r"^https?://", candidate, flags=re.I):
        candidate = "https://" + candidate
    parsed = urlparse(candidate)
    host = parsed.hostname or ""
    host_l = host.lower()
    signals: list[str] = []
    score = 0.08
    shorteners = {"bit.ly", "tinyurl.com", "t.co", "cutt.ly", "wa.link", "shorturl.at"}
    brands = {"sbi", "hdfc", "icici", "axisbank", "paytm", "phonepe", "gpay", "rbi", "amazon"}
    official = {"sbi.co.in", "hdfcbank.com", "icicibank.com", "axisbank.com", "paytm.com", "phonepe.com", "rbi.org.in", "amazon.in"}
    if host_l in shorteners:
        score += 0.22
        signals.append("url_shortener")
    if any(brand in host_l for brand in brands) and host_l not in official and not any(host_l.endswith("." + item) for item in official):
        score += 0.45
        signals.append("url_brand_spoof")
    if any(term in host_l for term in ("verify", "kyc", "support", "refund", "secure", "login")):
        score += 0.18
        signals.append("url_suspicious_keyword")
    if host_l.count("-") >= 2:
        score += 0.08
        signals.append("url_many_hyphens")
    if context_text and re.search(r"otp|kyc|blocked|urgent|verify", context_text, flags=re.I):
        score += 0.12
        signals.append("url_risky_context")
    score = min(0.98, score)
    type_score = round(score, 4)
    return {
        "modelVersion": URL_MODEL_VERSION,
        "score": type_score,
        "confidence": round(min(0.92, 0.5 + score * 0.45), 4),
        "scamTypeScores": {"PHISHING": type_score},
        "signals": signals,
    }


def build_model(dataset_path: Path) -> NaiveBayesScamModel | MiniLMCalibratedScamModel:
    samples = load_samples(dataset_path)
    requested = backend_requested()
    if requested != "naive_bayes":
        try:
            model = MiniLMCalibratedScamModel()
            model.train(samples)
            return model
        except Exception as exc:
            if requested == "minilm" and require_minilm():
                raise
            fallback = NaiveBayesScamModel(fallback_reason=f"MiniLM unavailable: {exc}")
            fallback.train(samples)
            return fallback

    model = NaiveBayesScamModel(fallback_reason="SCAMSHIELD_ML_BACKEND=naive_bayes")
    model.train(samples)
    return model
