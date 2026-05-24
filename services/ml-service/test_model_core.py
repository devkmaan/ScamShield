from __future__ import annotations

import os
import unittest
from pathlib import Path
from unittest.mock import patch

import model_core


ROOT = Path(__file__).resolve().parents[2]
DATASET_PATH = ROOT / "data" / "ml_training_samples.jsonl"


class ModelCoreTests(unittest.TestCase):
    def test_forced_naive_bayes_backend_keeps_score_contract(self) -> None:
        with patch.dict(os.environ, {"SCAMSHIELD_ML_BACKEND": "naive_bayes"}, clear=False):
            model = model_core.build_model(DATASET_PATH)

        result = model.predict_text("Your SBI KYC is blocked, click link and share OTP now")

        self.assertEqual(model.backend_name, "naive_bayes")
        self.assertEqual(result["modelVersion"], "text-scam-nb-v1")
        self.assertIn("calibrationVersion", result)
        self.assertGreaterEqual(result["score"], 0)
        self.assertLessEqual(result["score"], 1)
        self.assertTrue(result["scamTypeScores"])
        self.assertTrue(result["signals"])

    def test_auto_backend_falls_back_when_minilm_is_unavailable(self) -> None:
        with patch.dict(os.environ, {"SCAMSHIELD_ML_BACKEND": "auto", "SCAMSHIELD_REQUIRE_MINILM": ""}, clear=False):
            with patch.object(model_core.MiniLMCalibratedScamModel, "train", side_effect=RuntimeError("mock missing model")):
                model = model_core.build_model(DATASET_PATH)

        self.assertEqual(model.backend_name, "naive_bayes")
        self.assertIn("MiniLM unavailable", model.fallback_reason)
        self.assertIn("mock missing model", model.fallback_reason)

    def test_required_minilm_mode_raises_on_startup_failure(self) -> None:
        with patch.dict(os.environ, {"SCAMSHIELD_ML_BACKEND": "minilm", "SCAMSHIELD_REQUIRE_MINILM": "true"}, clear=False):
            with patch.object(model_core.MiniLMCalibratedScamModel, "train", side_effect=RuntimeError("mock missing model")):
                with self.assertRaises(RuntimeError):
                    model_core.build_model(DATASET_PATH)

    def test_url_model_contract(self) -> None:
        result = model_core.score_url("https://sbi-verify-support-login.example/kyc", "urgent kyc otp")

        self.assertEqual(result["modelVersion"], "url-lexical-v1")
        self.assertGreater(result["score"], 0.5)
        self.assertIn("PHISHING", result["scamTypeScores"])
        self.assertIn("url_brand_spoof", result["signals"])


if __name__ == "__main__":
    unittest.main()
