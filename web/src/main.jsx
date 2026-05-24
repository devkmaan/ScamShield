import React, { createContext, useContext, useEffect, useMemo, useState } from 'react';
import { createRoot } from 'react-dom/client';
import {
  Activity,
  AlertTriangle,
  BarChart3,
  Bot,
  BrainCircuit,
  CheckCircle2,
  ChevronLeft,
  ChevronRight,
  ClipboardCheck,
  ClipboardList,
  Clock3,
  Copy,
  Database,
  ExternalLink,
  Eye,
  Filter,
  FileText,
  FileWarning,
  History,
  Info,
  Languages,
  Link2,
  Loader2,
  MessageSquare,
  MessageSquareWarning,
  Play,
  QrCode,
  RefreshCw,
  Search,
  Send,
  ShieldAlert,
  ShieldCheck,
  Sparkles,
  Store,
  TrendingUp,
  Upload,
  WalletCards,
  X
} from 'lucide-react';
import { api, jsonBody } from './api.js';
import './styles.css';

const TABS = [
  { id: 'dashboard', labelKey: 'nav.dashboard', icon: BarChart3 },
  { id: 'check', labelKey: 'nav.check', icon: Search },
  { id: 'whatsapp', labelKey: 'nav.whatsapp', icon: Bot },
  { id: 'recovery', labelKey: 'nav.recovery', icon: ClipboardList },
  { id: 'merchant', labelKey: 'nav.merchant', icon: Store },
  { id: 'model', labelKey: 'nav.model', icon: BrainCircuit },
  { id: 'events', labelKey: 'nav.events', icon: Activity }
];

const LANGUAGES = [
  { code: 'en', label: 'English' },
  { code: 'hinglish', label: 'Hinglish' },
  { code: 'hi', label: 'Hindi' },
  { code: 'bn', label: 'Bengali' },
  { code: 'ta', label: 'Tamil' },
  { code: 'te', label: 'Telugu' },
  { code: 'mr', label: 'Marathi' },
  { code: 'gu', label: 'Gujarati' },
  { code: 'kn', label: 'Kannada' },
  { code: 'ml', label: 'Malayalam' },
  { code: 'pa', label: 'Punjabi' },
  { code: 'ur', label: 'Urdu' }
];

const INPUT_MODES = [
  { id: 'TEXT', labelKey: 'mode.message', icon: MessageSquare },
  { id: 'URL', labelKey: 'mode.link', icon: Link2 },
  { id: 'UPI_ID', labelKey: 'mode.upi', icon: WalletCards },
  { id: 'QR', labelKey: 'mode.qr', icon: QrCode },
  { id: 'SCREENSHOT', labelKey: 'mode.screenshot', icon: Upload }
];

const SAMPLE_CASES = [
  {
    labelKey: 'sample.kyc',
    inputType: 'TEXT',
    text: 'Your SBI KYC is blocked. Click https://sbi-verify-support.com/login and share OTP immediately.'
  },
  {
    labelKey: 'sample.qr',
    inputType: 'QR',
    text: 'upi://pay?pa=refund-care@okaxis&pn=Marketplace%20Refund%20Agent&am=4999&tn=Refund receive money by entering UPI PIN'
  },
  {
    labelKey: 'sample.job',
    inputType: 'TEXT',
    text: 'Part time job daily task. Pay registration fee first and earn high commission after each task.'
  },
  {
    labelKey: 'sample.investment',
    inputType: 'TEXT',
    text: 'Guaranteed crypto profit. Double your money in 7 days with private stock tips.'
  },
  {
    labelKey: 'sample.safe',
    inputType: 'URL',
    text: 'https://www.hdfcbank.com/personal/pay/cards'
  }
];

const SCAM_TYPES = ['UPI_COLLECT', 'PHISHING', 'IMPERSONATION', 'JOB_SCAM', 'INVESTMENT', 'LOAN_APP', 'FAKE_RECEIPT', 'UNKNOWN'];
const RISK_LEVELS = ['LOW', 'CAUTION', 'HIGH_RISK', 'CRITICAL'];

const DEFAULT_COPY = {
  'app.eyebrow': 'Consumer scam shield',
  'app.title': 'ScamShield command center',
  'app.ready': 'ready',
  'app.checking': 'checking',
  'app.backendUnavailable': 'backend unavailable',
  'app.user': 'User',
  'nav.dashboard': 'Overview',
  'nav.check': 'Check',
  'nav.whatsapp': 'WhatsApp',
  'nav.recovery': 'Recovery',
  'nav.merchant': 'Payees',
  'nav.model': 'Models',
  'nav.events': 'Events',
  'dashboard.eyebrow': 'Live protection',
  'dashboard.title': 'Risk operations',
  'dashboard.generate': 'Generate demo data',
  'dashboard.generating': 'Generating',
  'dashboard.scamMix': 'Scam mix',
  'dashboard.riskMix': 'Risk mix',
  'dashboard.recentActivity': 'Activity',
  'dashboard.reviewQueue': 'Review queue',
  'dashboard.analytics': 'Analytics mode',
  'dashboard.decisionDetail': 'Decision detail',
  'metric.decisions': 'Decisions',
  'metric.highRisk': 'High risk',
  'metric.reviewQueue': 'Review queue',
  'metric.evidence': 'Evidence',
  'metric.highRate': 'High-risk rate',
  'panel.recentDecisions': 'Recent decisions',
  'panel.topRiskPayees': 'Top risk payees',
  'panel.output': 'Result',
  'panel.eventLog': 'Event timeline',
  'check.title': 'Guided risk check',
  'check.samples': 'Sample cases',
  'check.history': 'Recent checks',
  'check.result': 'Decision',
  'check.noSelection': 'No decision selected',
  'field.inputType': 'Input type',
  'field.content': 'Content',
  'field.language': 'Language',
  'field.phone': 'Phone',
  'field.userId': 'User ID',
  'field.scamType': 'Scam type',
  'field.riskLevel': 'Risk level',
  'field.reviewOnly': 'Review only',
  'button.analyze': 'Analyze',
  'button.analyzing': 'Analyzing',
  'button.refresh': 'Refresh',
  'button.send': 'Send',
  'button.setLanguage': 'Set language',
  'button.outbox': 'Outbox',
  'button.createDraft': 'Create draft',
  'button.saveEvidence': 'Save evidence',
  'button.observe': 'Observe',
  'button.report': 'Report',
  'button.score': 'Score',
  'button.metadata': 'Metadata',
  'button.copy': 'Copy',
  'button.share': 'Share summary',
  'button.feedbackScam': 'Scam',
  'button.feedbackSafe': 'Not scam',
  'button.needHelp': 'Need help',
  'button.openReport': 'Open report',
  'button.useSample': 'Use sample',
  'button.resetAnalytics': 'Reset analytics',
  'button.previous': 'Previous',
  'button.next': 'Next',
  'button.clear': 'Clear',
  'option.allScams': 'All scams',
  'option.allRisks': 'All risks',
  'option.allTypes': 'All input types',
  'option.allLanguages': 'All languages',
  'option.reviewAny': 'All decisions',
  'option.reviewOnly': 'Needs review only',
  'analytics.totalMatches': 'matching decisions',
  'analytics.pageStatus': 'Page',
  'analytics.noMatches': 'No decisions match the current filters.',
  'analytics.noSelection': 'Select a decision to view its details.',
  'analytics.activeFilters': 'Active filters',
  'analytics.hasReport': 'Report',
  'analytics.hasPayee': 'Payee',
  'analytics.needsReview': 'Review',
  'mode.message': 'Message',
  'mode.link': 'Link',
  'mode.upi': 'UPI ID',
  'mode.qr': 'QR',
  'mode.screenshot': 'Screenshot',
  'sample.kyc': 'KYC OTP',
  'sample.qr': 'UPI QR',
  'sample.job': 'Job task',
  'sample.investment': 'Investment',
  'sample.safe': 'Safe merchant',
  'decision.score': 'Score',
  'decision.confidence': 'Confidence',
  'decision.type': 'Type',
  'decision.signals': 'Signals',
  'decision.actions': 'Actions',
  'decision.officialHelp': 'Official help',
  'decision.share': 'Share-safe summary',
  'decision.model': 'Model trace',
  'decision.report': 'Recovery checklist',
  'whatsapp.title': 'WhatsApp preview',
  'whatsapp.message': 'Forwarded message',
  'recovery.title': 'Recovery case',
  'recovery.lossContext': 'Loss context',
  'merchant.title': 'Merchant and payee risk',
  'merchant.upiId': 'UPI ID',
  'merchant.alias': 'Alias',
  'model.title': 'Model lab',
  'model.text': 'Text',
  'model.health': 'Service health',
  'model.output': 'Model output',
  'empty.noDecision': 'No decision yet.',
  'empty.noDecisions': 'No decisions yet.',
  'empty.noPayees': 'No payees yet.',
  'status.waiting': 'waiting',
  'table.risk': 'Risk',
  'table.score': 'Score',
  'table.type': 'Type',
  'table.model': 'Model',
  'table.signals': 'Signals',
  'table.complaints': 'Complaints',
  'table.payeeHash': 'Payee hash',
  'details.developer': 'Developer details'
};

const INFO_COPY = {
  'tip.metric.decisions': 'Total decisions stored by the local backend for this demo session.',
  'tip.metric.highRisk': 'Decisions classified as HIGH_RISK or CRITICAL by rules, ML, or merchant graph signals.',
  'tip.metric.reviewQueue': 'Decisions that need human review because confidence, risk, or graph signals are uncertain.',
  'tip.metric.highRate': 'Share of all stored decisions that are HIGH_RISK or CRITICAL.',
  'tip.dashboard.scamMix': 'Distribution of recent decisions by detected scam category. Click a row to filter the decision list.',
  'tip.dashboard.riskMix': 'Distribution of recent decisions by verdict level. Click a row to filter the decision list.',
  'tip.dashboard.analytics': 'Filter, paginate, and inspect stored decisions without exposing raw private evidence.',
  'tip.panel.recentDecisions': 'Most recent fraud checks stored by the local backend, filterable by risk and scam type.',
  'tip.panel.topRiskPayees': 'Hashed payee or merchant identifiers ranked by complaint count, risk score, alias changes, and connected risky identifiers. Raw UPI IDs are not shown for privacy.',
  'tip.dashboard.decisionDetail': 'A privacy-safe detail view for the selected decision, including verdict facts, model trace, actions, and share summary.',
  'tip.service.api': 'Go risk orchestrator and public API status.',
  'tip.service.ml': 'Text and URL model-serving status used for flexible fraud scoring.',
  'tip.service.genai': 'Local GenAI renderer and normalizer status. It explains decisions but does not decide risk.',
  'tip.check.title': 'Run a guided scam check for messages, links, UPI IDs, QR payloads, or screenshot references.',
  'tip.check.history': 'Recent checks for the selected user, merged from local storage and backend history.',
  'tip.check.result': 'The final risk decision created by the Go orchestrator.',
  'tip.field.content': 'Paste only the suspicious content needed for analysis. Sensitive values are redacted by the backend.',
  'tip.field.language': 'Controls the preferred language for user-facing replies and GenAI rendering.',
  'tip.field.userId': 'Local demo user identity used to group history and preferences.',
  'tip.field.phone': 'WhatsApp simulator sender number used to group bot replies.',
  'tip.whatsapp.message': 'Message body forwarded into the mocked WhatsApp webhook.',
  'tip.whatsapp.title': 'Simulates WhatsApp webhook messages and bot replies without using the real WhatsApp Cloud API.',
  'tip.whatsapp.output': 'Chat-style preview of user messages and generated bot replies.',
  'tip.recovery.title': 'Creates a draft recovery checklist for users who already paid or shared sensitive information.',
  'tip.recovery.lossContext': 'Brief description of what happened, amount, channel, or transaction clues for recovery guidance.',
  'tip.decision.report': 'Step-by-step recovery guidance with official 1930 and cybercrime.gov.in reporting paths.',
  'tip.merchant.title': 'Observe or report a payee so the merchant graph can update risk features.',
  'tip.merchant.upiId': 'Raw UPI/payee identifier entered locally; backend hashes it before storing risk records.',
  'tip.merchant.alias': 'Display name or label seen for the payee, used only as a risk feature.',
  'tip.model.title': 'Inspect model scoring and service metadata for the current local setup.',
  'tip.model.text': 'Text sent directly to the model service to inspect classifier output.',
  'tip.model.health': 'Service health and active model/version information.',
  'tip.panel.eventLog': 'Recent backend events such as risk decisions, WhatsApp replies, feedback, and evidence actions.',
  'tip.decision.score': 'Risk probability-style score from rules, ML, URL, and merchant graph signals.',
  'tip.decision.confidence': 'How strongly the system trusts the available evidence for this verdict.',
  'tip.decision.signals': 'Top machine-readable reasons that contributed to the risk decision.',
  'tip.decision.actions': 'User-safe next steps derived from immutable risk facts.',
  'tip.decision.officialHelp': 'Official Indian cybercrime reporting and support references.',
  'tip.decision.model': 'Rules, ML, merchant graph, and GenAI versions involved in the decision.',
  'tip.decision.share': 'A privacy-safe summary designed to copy or share without raw UPI IDs or private evidence.'
};

const COPY_KEYS = { ...DEFAULT_COPY, ...INFO_COPY };

const AppContext = createContext({
  language: 'en',
  setLanguage: () => {},
  userId: 'react-user',
  setUserId: () => {},
  t: (key, fallback) => fallback || COPY_KEYS[key] || key
});

function App() {
  const [active, setActive] = useState('dashboard');
  const [ready, setReady] = useState(null);
  const [summary, setSummary] = useState(null);
  const [insights, setInsights] = useState(null);
  const [events, setEvents] = useState([]);
  const [modelMeta, setModelMeta] = useState(null);
  const [genaiMeta, setGenaiMeta] = useState(null);
  const [error, setError] = useState('');
  const [language, setLanguageState] = useState(() => localStorage.getItem('scamshield.language') || 'en');
  const [userId, setUserIdState] = useState(() => localStorage.getItem('scamshield.userId') || 'react-user');
  const [copy, setCopy] = useState(COPY_KEYS);

  function setLanguage(nextLanguage) {
    setLanguageState(nextLanguage);
    localStorage.setItem('scamshield.language', nextLanguage);
  }

  function setUserId(nextUserId) {
    setUserIdState(nextUserId);
    localStorage.setItem('scamshield.userId', nextUserId);
  }

  useEffect(() => {
    let cancelled = false;
    async function loadBundle() {
      try {
        const cached = localStorage.getItem(`scamshield.i18n.${language}`);
        if (cached) setCopy({ ...COPY_KEYS, ...JSON.parse(cached) });
        const payload = await api('/v1/i18n/bundle', jsonBody({ language, keys: COPY_KEYS }));
        const bundle = payload.bundle || {};
        if (!cancelled) {
          setCopy({ ...COPY_KEYS, ...bundle });
          localStorage.setItem(`scamshield.i18n.${language}`, JSON.stringify(bundle));
        }
      } catch {
        if (!cancelled) setCopy(COPY_KEYS);
      }
    }
    loadBundle();
    return () => {
      cancelled = true;
    };
  }, [language]);

  async function refresh() {
    try {
      setError('');
      const [readyPayload, summaryPayload, insightsPayload, mlPayload, genaiPayload] = await Promise.all([
        api('/ready'),
        api('/v1/admin/summary'),
        api('/v1/insights/trends'),
        api('/internal/model/metadata'),
        api('/internal/genai/metadata')
      ]);
      setReady(readyPayload);
      setSummary(summaryPayload);
      setInsights(insightsPayload);
      setModelMeta(mlPayload);
      setGenaiMeta(genaiPayload);
      if (active === 'events') {
        const eventPayload = await api('/v1/events');
        setEvents(eventPayload.items || []);
      }
    } catch (err) {
      setError(err.message);
    }
  }

  useEffect(() => {
    refresh();
    const id = setInterval(refresh, 10000);
    return () => clearInterval(id);
  }, [active]);

  const context = useMemo(() => ({
    language,
    setLanguage,
    userId,
    setUserId,
    t: (key, fallback) => copy[key] || COPY_KEYS[key] || fallback || key
  }), [copy, language, userId]);

  return (
    <AppContext.Provider value={context}>
      <div className="appShell">
        <aside className="sidebar">
          <div className="brandBlock">
            <div className="brandMark"><ShieldAlert size={22} /></div>
            <div>
              <strong>ScamShield</strong>
              <span>{context.t('app.eyebrow')}</span>
            </div>
          </div>
          <nav className="navRail">
            {TABS.map((tab) => {
              const Icon = tab.icon;
              const label = context.t(tab.labelKey);
              return (
                <button key={tab.id} className={active === tab.id ? 'active' : ''} onClick={() => setActive(tab.id)} title={label}>
                  <Icon size={18} />
                  <span>{label}</span>
                </button>
              );
            })}
          </nav>
          <div className="sidebarFooter">
            <StatusDot ok={!error && ready?.status === 'ready'} label={error ? context.t('app.backendUnavailable') : ready?.status || context.t('app.checking')} />
          </div>
        </aside>
        <div className="workspace">
          <Header ready={ready} error={error} onRefresh={refresh} />
          <main>
            {error && <Notice text={error} />}
            {active === 'dashboard' && <Dashboard summary={summary} insights={insights} ready={ready} modelMeta={modelMeta} genaiMeta={genaiMeta} onSeed={refresh} />}
            {active === 'check' && <RiskCheck onDone={refresh} />}
            {active === 'whatsapp' && <WhatsAppPanel onDone={refresh} />}
            {active === 'recovery' && <RecoveryPanel onDone={refresh} />}
            {active === 'merchant' && <MerchantPanel onDone={refresh} summary={summary} />}
            {active === 'model' && <ModelLab modelMeta={modelMeta} genaiMeta={genaiMeta} />}
            {active === 'events' && <EventsPanel events={events} onRefresh={refresh} />}
          </main>
        </div>
      </div>
    </AppContext.Provider>
  );
}

function useApp() {
  return useContext(AppContext);
}

function Header({ ready, error, onRefresh }) {
  const { language, setLanguage, userId, setUserId, t } = useApp();
  return (
    <header className="topbar">
      <div>
        <p className="eyebrow">{t('app.eyebrow')}</p>
        <h1>{t('app.title')}</h1>
      </div>
      <div className="topbarActions">
        <label className="compactField" title={t('app.user')}>
          <span>{t('app.user')}<InfoTip textKey="tip.field.userId" /></span>
          <input value={userId} onChange={(event) => setUserId(event.target.value)} />
        </label>
        <label className="compactField" title={t('field.language')}>
          <span><Languages size={14} />{t('field.language')}<InfoTip textKey="tip.field.language" /></span>
          <select value={language} onChange={(event) => setLanguage(event.target.value)}>
            {LANGUAGES.map((item) => <option key={item.code} value={item.code}>{item.label}</option>)}
          </select>
        </label>
        <StatusDot ok={!error && ready?.status === 'ready'} label={error ? t('app.backendUnavailable') : ready?.status || t('app.checking')} />
        <IconButton label={t('button.refresh')} onClick={onRefresh} icon={RefreshCw} />
      </div>
    </header>
  );
}

function Dashboard({ summary, insights, ready, modelMeta, genaiMeta, onSeed }) {
  const { t } = useApp();
  const [busy, setBusy] = useState(false);
  const [filters, setFilters] = useState({ scamType: '', riskLevel: '', inputType: '', language: '', needsHumanReview: '' });
  const [page, setPage] = useState(1);
  const [decisionPage, setDecisionPage] = useState({ items: [], page: 1, pageSize: 10, total: 0, totalPages: 0, hasNext: false, hasPrevious: false });
  const [loadingDecisions, setLoadingDecisions] = useState(false);
  const [selectedDecision, setSelectedDecision] = useState(null);
  const [share, setShare] = useState(null);
  const highRiskRate = summary?.decisionCount ? Math.round((summary.highRiskCount / summary.decisionCount) * 100) : 0;

  useEffect(() => {
    loadDecisionPage();
  }, [filters, page, summary?.decisionCount]);

  async function seed() {
    setBusy(true);
    await api('/v1/simulate/stream', jsonBody({ count: 12 }));
    setBusy(false);
    onSeed();
  }

  async function loadDecisionPage() {
    setLoadingDecisions(true);
    try {
      const params = new URLSearchParams({ page: String(page), pageSize: '10' });
      Object.entries(filters).forEach(([key, value]) => {
        if (value) params.set(key, value);
      });
      const payload = await api(`/v1/admin/decisions?${params.toString()}`);
      setDecisionPage(payload);
      if (selectedDecision && !matchesDecisionFilters(selectedDecision, filters)) {
        setSelectedDecision(null);
        setShare(null);
      }
    } finally {
      setLoadingDecisions(false);
    }
  }

  function updateFilter(key, value) {
    setFilters((current) => ({ ...current, [key]: value }));
    setPage(1);
  }

  function resetFilters() {
    setFilters({ scamType: '', riskLevel: '', inputType: '', language: '', needsHumanReview: '' });
    setPage(1);
  }

  async function selectDecision(decisionId) {
    const decision = await api(`/v1/decisions/${decisionId}`);
    setSelectedDecision(decision);
    setShare(null);
  }

  async function loadShare() {
    if (!selectedDecision?.decisionId) return;
    const payload = await api(`/v1/decisions/${selectedDecision.decisionId}/share`);
    setShare(payload);
    await copyToClipboard(payload.shareText);
  }

  return (
    <section className="pageStack">
      <div className="pageTitle">
        <div>
          <p className="eyebrow">{t('dashboard.eyebrow')}</p>
          <h2>{t('dashboard.title')}</h2>
        </div>
        <button className="primary" onClick={seed} disabled={busy}>
          {busy ? <Loader2 size={17} className="spin" /> : <Play size={17} />}
          <span>{busy ? t('dashboard.generating') : t('dashboard.generate')}</span>
        </button>
      </div>

      <div className="metricGrid">
        <Metric label={t('metric.decisions')} value={summary?.decisionCount || 0} icon={Database} infoKey="tip.metric.decisions" />
        <Metric label={t('metric.highRisk')} value={summary?.highRiskCount || 0} icon={AlertTriangle} tone="danger" infoKey="tip.metric.highRisk" />
        <Metric label={t('metric.reviewQueue')} value={summary?.humanReviewCount || 0} icon={FileWarning} tone="warn" infoKey="tip.metric.reviewQueue" />
        <Metric label={t('metric.highRate')} value={`${highRiskRate}%`} icon={TrendingUp} tone="info" infoKey="tip.metric.highRate" />
      </div>

      <ServiceStrip ready={ready} modelMeta={modelMeta} genaiMeta={genaiMeta} />

      <div className="dashboardGrid">
        <Panel title={t('dashboard.scamMix')} infoKey="tip.dashboard.scamMix" action={<Sparkles size={18} />}>
          <BucketBars items={insights?.scamTypes || []} selected={filters.scamType} onSelect={(label) => updateFilter('scamType', label)} />
        </Panel>
        <Panel title={t('dashboard.riskMix')} infoKey="tip.dashboard.riskMix" action={<ShieldCheck size={18} />}>
          <BucketBars items={insights?.riskLevels || []} selected={filters.riskLevel} onSelect={(label) => updateFilter('riskLevel', label)} />
        </Panel>
      </div>

      <Panel title={t('dashboard.analytics')} infoKey="tip.dashboard.analytics" action={<Filter size={18} />}>
        <AnalyticsFilters filters={filters} onChange={updateFilter} onReset={resetFilters} />
        <ActiveFilterChips filters={filters} onClear={(key) => updateFilter(key, '')} />
      </Panel>

      <div className="analyticsGrid">
        <Panel title={t('panel.recentDecisions')} infoKey="tip.panel.recentDecisions" action={loadingDecisions ? <Loader2 size={18} className="spin" /> : <History size={18} />}>
          <DecisionList
            items={decisionPage.items || []}
            selectedId={selectedDecision?.decisionId}
            onSelect={selectDecision}
            emptyText={t('analytics.noMatches')}
          />
          <Pagination page={decisionPage} onPage={setPage} />
        </Panel>
        <Panel title={t('dashboard.decisionDetail')} infoKey="tip.dashboard.decisionDetail" action={selectedDecision ? <RiskBadge level={selectedDecision.riskLevel} /> : <Clock3 size={18} />}>
          {selectedDecision ? (
            <DecisionDetail decision={selectedDecision} share={share} onShare={loadShare} />
          ) : (
            <EmptyState icon={FileText} text={t('analytics.noSelection')} />
          )}
        </Panel>
      </div>

      <div className="dashboardGrid">
        <Panel title={t('panel.topRiskPayees')} infoKey="tip.panel.topRiskPayees">
          <MerchantList items={summary?.topRiskMerchants || []} />
        </Panel>
      </div>
    </section>
  );
}

function RiskCheck({ onDone }) {
  const { language, userId, t } = useApp();
  const [inputType, setInputType] = useState('TEXT');
  const [text, setText] = useState(SAMPLE_CASES[0].text);
  const [result, setResult] = useState(null);
  const [report, setReport] = useState(null);
  const [share, setShare] = useState(null);
  const [history, setHistory] = useState([]);
  const [loading, setLoading] = useState(false);
  const [feedback, setFeedback] = useState('');

  useEffect(() => {
    loadHistory();
  }, [userId]);

  async function loadHistory() {
    const local = loadLocalHistory(userId);
    setHistory(local);
    try {
      const payload = await api(`/v1/users/${encodeURIComponent(userId)}/history?limit=20`);
      setHistory(mergeDecisions(payload.items || [], local));
    } catch {
      setHistory(local);
    }
  }

  async function submit() {
    setLoading(true);
    setFeedback('');
    setShare(null);
    setReport(null);
    try {
      const body = { inputType, userId, language };
      if (inputType === 'URL') body.url = text;
      else if (inputType === 'UPI_ID') body.upiId = text;
      else if (inputType === 'QR') body.qrPayload = text;
      else if (inputType === 'SCREENSHOT') body.mediaRef = text;
      else body.text = text;
      const decision = await api('/v1/check', jsonBody(body));
      setResult(decision);
      saveLocalHistory(userId, decision);
      setHistory((items) => mergeDecisions([decision], items));
      if (decision.reportId) setReport(await api(`/v1/reports/${decision.reportId}`));
      onDone();
    } finally {
      setLoading(false);
    }
  }

  async function selectDecision(decisionId) {
    const decision = await api(`/v1/decisions/${decisionId}`);
    setResult(decision);
    setShare(null);
    setFeedback('');
    setReport(decision.reportId ? await api(`/v1/reports/${decision.reportId}`) : null);
  }

  async function loadShare(decision = result) {
    if (!decision?.decisionId) return;
    const payload = await api(`/v1/decisions/${decision.decisionId}/share`);
    setShare(payload);
    await copyToClipboard(payload.shareText);
  }

  async function sendFeedback(verdict) {
    if (!result?.decisionId) return;
    await api('/v1/feedback', jsonBody({ decisionId: result.decisionId, userId, verdict }));
    setFeedback(verdict);
  }

  return (
    <section className="checkLayout">
      <div className="workflowColumn">
        <Panel title={t('check.title')} infoKey="tip.check.title" action={<Search size={18} />}>
          <SegmentedControl items={INPUT_MODES} value={inputType} onChange={setInputType} />
          <div className="sampleRow">
            {SAMPLE_CASES.map((sample) => (
              <button key={sample.labelKey} className="chipButton" onClick={() => {
                setInputType(sample.inputType);
                setText(sample.text);
              }}>
                <Sparkles size={14} />
                <span>{t(sample.labelKey)}</span>
              </button>
            ))}
          </div>
          <Field label={t('field.content')} infoKey="tip.field.content">
            <textarea className="largeInput" value={text} onChange={(event) => setText(event.target.value)} />
          </Field>
          <div className="buttonRow">
            <button className="primary" onClick={submit} disabled={loading}>
              {loading ? <Loader2 size={17} className="spin" /> : <Search size={17} />}
              <span>{loading ? t('button.analyzing') : t('button.analyze')}</span>
            </button>
          </div>
        </Panel>
        <Panel title={t('check.history')} infoKey="tip.check.history" action={<History size={18} />}>
          <HistoryList items={history} selectedId={result?.decisionId} onSelect={selectDecision} />
        </Panel>
      </div>
      <div className="resultColumn">
        <Panel title={t('check.result')} infoKey="tip.check.result" action={result ? <RiskBadge level={result.riskLevel} /> : <Clock3 size={18} />}>
          {result ? (
            <DecisionCard
              decision={result}
              report={report}
              share={share}
              feedback={feedback}
              onShare={() => loadShare(result)}
              onFeedback={sendFeedback}
            />
          ) : (
            <EmptyState icon={ShieldCheck} text={t('empty.noDecision')} />
          )}
        </Panel>
      </div>
    </section>
  );
}

function WhatsAppPanel({ onDone }) {
  const { language, t } = useApp();
  const [phone, setPhone] = useState('919999999999');
  const [message, setMessage] = useState('KYC update urgent. Click https://paytm-verify-help.com and share OTP');
  const [conversation, setConversation] = useState([]);

  async function sendText(body) {
    setConversation((items) => [...items, { from: 'user', text: body, createdAt: new Date().toISOString() }]);
    const payload = {
      entry: [{
        changes: [{
          value: {
            messages: [{
              id: `wamid.react.${Date.now()}`,
              from: phone,
              type: 'text',
              text: { body }
            }]
          }
        }]
      }]
    };
    await api('/webhooks/whatsapp', jsonBody(payload));
    setTimeout(load, 450);
    onDone();
  }

  async function load() {
    const payload = await api(`/v1/outbox?userId=${encodeURIComponent(phone)}`);
    const replies = (payload.items || []).map((item) => ({ from: 'bot', text: item.text, createdAt: item.createdAt, buttons: item.buttons }));
    setConversation((items) => mergeConversation(items.filter((item) => item.from === 'user'), replies));
  }

  return (
    <section className="twoColumn">
      <Panel title={t('whatsapp.title')} infoKey="tip.whatsapp.title" action={<Bot size={18} />}>
        <Field label={t('field.phone')} infoKey="tip.field.phone"><input value={phone} onChange={(event) => setPhone(event.target.value)} /></Field>
        <Field label={t('whatsapp.message')} infoKey="tip.whatsapp.message"><textarea value={message} onChange={(event) => setMessage(event.target.value)} /></Field>
        <div className="buttonRow">
          <button className="primary" onClick={() => sendText(message)}><Send size={17} /><span>{t('button.send')}</span></button>
          <button onClick={() => sendText(`/language ${language}`)}><Languages size={17} /><span>{t('button.setLanguage')}</span></button>
          <button onClick={load}><RefreshCw size={17} /><span>{t('button.outbox')}</span></button>
        </div>
      </Panel>
      <Panel title={t('panel.output')} infoKey="tip.whatsapp.output" action={<MessageSquareWarning size={18} />}>
        <ChatPreview items={conversation} />
      </Panel>
    </section>
  );
}

function RecoveryPanel({ onDone }) {
  const { language, userId, t } = useApp();
  const [text, setText] = useState('I already paid 5000 after a Telegram task commission message to taskbonus@paytm.');
  const [decision, setDecision] = useState(null);
  const [report, setReport] = useState(null);
  const [share, setShare] = useState(null);

  async function start() {
    setShare(null);
    const data = await api('/v1/recovery/start', jsonBody({ userId, inputType: 'TEXT', text, language }));
    setDecision(data);
    setReport(data.reportId ? await api(`/v1/reports/${data.reportId}`) : null);
    onDone();
  }

  async function evidence() {
    const data = await api('/v1/evidence', jsonBody({ userId, mediaType: 'TEXT', source: 'REACT_CONSOLE', content: text, decisionId: decision?.decisionId, reportId: decision?.reportId }));
    setReport((current) => current ? { ...current, lastEvidenceId: data.evidenceId } : current);
    onDone();
  }

  async function shareDecision() {
    if (!decision?.decisionId) return;
    const payload = await api(`/v1/decisions/${decision.decisionId}/share`);
    setShare(payload);
    await copyToClipboard(payload.shareText);
  }

  return (
    <section className="twoColumn">
      <Panel title={t('recovery.title')} infoKey="tip.recovery.title" action={<ClipboardList size={18} />}>
        <Field label={t('recovery.lossContext')} infoKey="tip.recovery.lossContext"><textarea value={text} onChange={(event) => setText(event.target.value)} /></Field>
        <div className="buttonRow">
          <button className="primary" onClick={start}><ClipboardCheck size={17} /><span>{t('button.createDraft')}</span></button>
          <button onClick={evidence}><Upload size={17} /><span>{t('button.saveEvidence')}</span></button>
        </div>
      </Panel>
      <Panel title={t('decision.report')} infoKey="tip.decision.report" action={<FileText size={18} />}>
        {report ? <RecoveryCard report={report} decision={decision} share={share} onShare={shareDecision} /> : <EmptyState icon={ClipboardList} text={t('status.waiting')} />}
      </Panel>
    </section>
  );
}

function MerchantPanel({ summary }) {
  const { t } = useApp();
  const [rawPayee, setRawPayee] = useState('refund-care@okaxis');
  const [alias, setAlias] = useState('Marketplace Refund Agent');
  const [result, setResult] = useState(null);

  async function observe() {
    setResult(await api('/internal/payee/observe', jsonBody({ rawPayee, alias, source: 'REACT_CONSOLE' })));
  }

  async function report() {
    setResult(await api('/internal/payee/report', jsonBody({ rawPayee, alias, comment: 'Reported from React console' })));
  }

  return (
    <section className="twoColumn">
      <Panel title={t('merchant.title')} infoKey="tip.merchant.title" action={<Store size={18} />}>
        <Field label={t('merchant.upiId')} infoKey="tip.merchant.upiId"><input value={rawPayee} onChange={(event) => setRawPayee(event.target.value)} /></Field>
        <Field label={t('merchant.alias')} infoKey="tip.merchant.alias"><input value={alias} onChange={(event) => setAlias(event.target.value)} /></Field>
        <div className="buttonRow">
          <button className="primary" onClick={observe}><Eye size={17} /><span>{t('button.observe')}</span></button>
          <button onClick={report}><MessageSquareWarning size={17} /><span>{t('button.report')}</span></button>
        </div>
        {result && <MerchantResult item={result} />}
      </Panel>
      <Panel title={t('panel.topRiskPayees')} infoKey="tip.panel.topRiskPayees" action={<WalletCards size={18} />}>
        <MerchantList items={summary?.topRiskMerchants || []} />
      </Panel>
    </section>
  );
}

function ModelLab({ modelMeta, genaiMeta }) {
  const { language, t } = useApp();
  const [text, setText] = useState('Part time job daily task. Pay registration fee and earn commission.');
  const [result, setResult] = useState(null);

  async function score() {
    setResult(await api('/internal/model/score-text', jsonBody({ text, languageHint: language })));
  }

  async function metadata() {
    const [ml, genai] = await Promise.all([api('/internal/model/metadata'), api('/internal/genai/metadata')]);
    setResult({ ml, genai });
  }

  return (
    <section className="twoColumn">
      <Panel title={t('model.title')} infoKey="tip.model.title" action={<BrainCircuit size={18} />}>
        <Field label={t('model.text')} infoKey="tip.model.text"><textarea value={text} onChange={(event) => setText(event.target.value)} /></Field>
        <div className="buttonRow">
          <button className="primary" onClick={score}><BrainCircuit size={17} /><span>{t('button.score')}</span></button>
          <button onClick={metadata}><Database size={17} /><span>{t('button.metadata')}</span></button>
        </div>
        {result && <DeveloperDetails value={result} />}
      </Panel>
      <Panel title={t('model.health')} infoKey="tip.model.health" action={<Activity size={18} />}>
        <ServiceStrip modelMeta={modelMeta} genaiMeta={genaiMeta} compact />
      </Panel>
    </section>
  );
}

function EventsPanel({ events, onRefresh }) {
  const { t } = useApp();
  return (
    <Panel title={t('panel.eventLog')} infoKey="tip.panel.eventLog" action={<button onClick={onRefresh}><RefreshCw size={17} /><span>{t('button.refresh')}</span></button>}>
      <Timeline items={events} />
    </Panel>
  );
}

function AnalyticsFilters({ filters, onChange, onReset }) {
  const { t } = useApp();
  return (
    <div className="analyticsFilters">
      <label>
        <span>{t('field.scamType')}</span>
        <select value={filters.scamType} onChange={(event) => onChange('scamType', event.target.value)}>
          <option value="">{t('option.allScams')}</option>
          {SCAM_TYPES.map((item) => <option key={item} value={item}>{item}</option>)}
        </select>
      </label>
      <label>
        <span>{t('field.riskLevel')}</span>
        <select value={filters.riskLevel} onChange={(event) => onChange('riskLevel', event.target.value)}>
          <option value="">{t('option.allRisks')}</option>
          {RISK_LEVELS.map((item) => <option key={item} value={item}>{item}</option>)}
        </select>
      </label>
      <label>
        <span>{t('field.inputType')}</span>
        <select value={filters.inputType} onChange={(event) => onChange('inputType', event.target.value)}>
          <option value="">{t('option.allTypes')}</option>
          {INPUT_MODES.map((item) => <option key={item.id} value={item.id}>{t(item.labelKey)}</option>)}
        </select>
      </label>
      <label>
        <span>{t('field.language')}</span>
        <select value={filters.language} onChange={(event) => onChange('language', event.target.value)}>
          <option value="">{t('option.allLanguages')}</option>
          {LANGUAGES.map((item) => <option key={item.code} value={item.code}>{item.label}</option>)}
        </select>
      </label>
      <label>
        <span>{t('field.reviewOnly')}</span>
        <select value={filters.needsHumanReview} onChange={(event) => onChange('needsHumanReview', event.target.value)}>
          <option value="">{t('option.reviewAny')}</option>
          <option value="true">{t('option.reviewOnly')}</option>
        </select>
      </label>
      <button onClick={onReset}><X size={17} /><span>{t('button.resetAnalytics')}</span></button>
    </div>
  );
}

function ActiveFilterChips({ filters, onClear }) {
  const { t } = useApp();
  const labels = {
    scamType: t('field.scamType'),
    riskLevel: t('field.riskLevel'),
    inputType: t('field.inputType'),
    language: t('field.language'),
    needsHumanReview: t('field.reviewOnly')
  };
  const active = Object.entries(filters).filter(([, value]) => value);
  if (!active.length) return null;
  return (
    <div className="filterChips" aria-label={t('analytics.activeFilters')}>
      {active.map(([key, value]) => (
        <button key={key} className="filterChip" onClick={() => onClear(key)}>
          <span>{labels[key]}: {value === 'true' ? t('option.reviewOnly') : value}</span>
          <X size={14} />
        </button>
      ))}
    </div>
  );
}

function Pagination({ page, onPage }) {
  const { t } = useApp();
  const totalPages = page?.totalPages || 0;
  const current = page?.page || 1;
  return (
    <div className="pagination">
      <span>{page?.total || 0} {t('analytics.totalMatches')}</span>
      <div>
        <button disabled={!page?.hasPrevious} onClick={() => onPage(current - 1)}><ChevronLeft size={17} /><span>{t('button.previous')}</span></button>
        <strong>{t('analytics.pageStatus')} {totalPages ? current : 0} / {totalPages}</strong>
        <button disabled={!page?.hasNext} onClick={() => onPage(current + 1)}><span>{t('button.next')}</span><ChevronRight size={17} /></button>
      </div>
    </div>
  );
}

function DecisionDetail({ decision, share, onShare }) {
  const { t } = useApp();
  return (
    <div className="decisionDetail">
      <div className="decisionHeader">
        <RiskBadge level={decision.riskLevel} />
        <div className="decisionMeta">
          <span>{decision.scamType}</span>
          <span>{formatDateTime(decision.createdAt)}</span>
        </div>
      </div>
      <div className="scorePanel">
        <ScoreMeter label={t('decision.score')} value={decision.score} infoKey="tip.decision.score" />
        <ScoreMeter label={t('decision.confidence')} value={decision.confidence} infoKey="tip.decision.confidence" />
      </div>
      <div className="factGrid analyticsFacts">
        <span>{t('field.inputType')} <b>{decision.inputType}</b></span>
        <span>{t('field.language')} <b>{decision.language || 'en'}</b></span>
        <span>{t('analytics.needsReview')} <b>{decision.needsHumanReview ? 'Yes' : 'No'}</b></span>
        <span>{t('analytics.hasReport')} <b>{decision.reportId || 'None'}</b></span>
      </div>
      <p className="decisionMessage">{decision.userMessage}</p>
      <SignalRow signals={decision.topSignals || []} />
      <ActionList title={t('decision.actions')} items={decision.recommendedActions || []} infoKey="tip.decision.actions" />
      <ActionList title={t('decision.officialHelp')} items={['1930', 'cybercrime.gov.in']} infoKey="tip.decision.officialHelp" />
      <ModelVersionGrid versions={decision.modelVersions || {}} />
      {decision.reportId && <code>{t('decision.report')}: {decision.reportId}</code>}
      <div className="buttonRow">
        <button className="primary" onClick={onShare}><Copy size={17} /><span>{t('button.share')}</span></button>
      </div>
      {share && <ShareBox share={share} />}
      <DeveloperDetails value={decision} />
    </div>
  );
}

function ModelVersionGrid({ versions }) {
  const { t } = useApp();
  const entries = Object.entries(versions);
  if (!entries.length) return null;
  return (
    <div className="modelTrace">
      <strong className="labelWithInfo">{t('decision.model')}<InfoTip textKey="tip.decision.model" /></strong>
      <div>
        {entries.map(([key, value]) => <code key={key}>{key}: {value}</code>)}
      </div>
    </div>
  );
}

function DecisionCard({ decision, report, share, feedback, onShare, onFeedback }) {
  const { t } = useApp();
  return (
    <div className="decisionCard">
      <div className="decisionHeader">
        <RiskBadge level={decision.riskLevel} />
        <div className="decisionMeta">
          <span>{decision.scamType}</span>
          <span>{formatTime(decision.createdAt)}</span>
        </div>
      </div>
      <div className="scorePanel">
        <ScoreMeter label={t('decision.score')} value={decision.score} infoKey="tip.decision.score" />
        <ScoreMeter label={t('decision.confidence')} value={decision.confidence} infoKey="tip.decision.confidence" />
      </div>
      <p className="decisionMessage">{decision.userMessage}</p>
      <SignalRow signals={decision.topSignals || []} />
      <ActionList title={t('decision.actions')} items={decision.recommendedActions || []} infoKey="tip.decision.actions" />
      <ActionList title={t('decision.officialHelp')} items={['1930', 'cybercrime.gov.in']} infoKey="tip.decision.officialHelp" />
      <ModelVersionGrid versions={decision.modelVersions || {}} />
      {report && <RecoveryCard report={report} compact />}
      {share && <ShareBox share={share} />}
      <div className="buttonRow">
        <button className="primary" onClick={onShare}><Copy size={17} /><span>{t('button.share')}</span></button>
        <button className={feedback === 'SCAM' ? 'selected' : ''} onClick={() => onFeedback('SCAM')}><AlertTriangle size={17} /><span>{t('button.feedbackScam')}</span></button>
        <button className={feedback === 'NOT_SCAM' ? 'selected' : ''} onClick={() => onFeedback('NOT_SCAM')}><CheckCircle2 size={17} /><span>{t('button.feedbackSafe')}</span></button>
        <button onClick={() => onFeedback('NEED_HELP')}><ClipboardList size={17} /><span>{t('button.needHelp')}</span></button>
      </div>
      <DeveloperDetails value={decision} />
    </div>
  );
}

function RecoveryCard({ report, decision, share, onShare, compact = false }) {
  const { t } = useApp();
  return (
    <div className={compact ? 'recoveryBlock compact' : 'recoveryBlock'}>
      <div className="recoveryHeader">
        <ClipboardCheck size={18} />
        <div>
          <strong>{report.status}</strong>
          <span>{report.reportId}</span>
        </div>
      </div>
      <p>{report.summary}</p>
      <ol className="checklist">
        {(report.checklist || []).map((item) => <li key={item}>{item}</li>)}
      </ol>
      {!compact && (
        <>
          <ActionList title={t('decision.officialHelp')} items={report.officialHelp || []} infoKey="tip.decision.officialHelp" />
          {decision && <RiskBadge level={decision.riskLevel} />}
          {onShare && <button className="primary" onClick={onShare}><Copy size={17} /><span>{t('button.share')}</span></button>}
          {share && <ShareBox share={share} />}
        </>
      )}
    </div>
  );
}

function ShareBox({ share }) {
  const { t } = useApp();
  return (
    <div className="shareBox">
      <div className="shareHeader"><Copy size={16} /><strong className="labelWithInfo">{t('decision.share')}<InfoTip textKey="tip.decision.share" /></strong></div>
      <p>{share.shareText}</p>
    </div>
  );
}

function ServiceStrip({ ready, modelMeta, genaiMeta, compact = false }) {
  return (
    <div className={compact ? 'serviceGrid compact' : 'serviceGrid'}>
      {ready && <ServiceTile icon={ShieldCheck} label="API" value={ready.status || 'ready'} ok={ready.status === 'ready'} infoKey="tip.service.api" />}
      <ServiceTile icon={BrainCircuit} label="ML" value={modelMeta?.activeModels?.text || modelMeta?.mode || 'model'} ok={Boolean(modelMeta)} infoKey="tip.service.ml" />
      <ServiceTile icon={Sparkles} label="GenAI" value={genaiMeta?.activeModels?.renderer || genaiMeta?.mode || 'renderer'} ok={Boolean(genaiMeta)} infoKey="tip.service.genai" />
    </div>
  );
}

function ServiceTile({ icon: Icon, label, value, ok, infoKey }) {
  return (
    <div className="serviceTile">
      <Icon size={18} />
      <div>
        <span className="labelWithInfo">{label}{infoKey && <InfoTip textKey={infoKey} />}</span>
        <strong>{value}</strong>
      </div>
      <StatusDot ok={ok} label={ok ? 'online' : 'offline'} compact />
    </div>
  );
}

function SegmentedControl({ items, value, onChange }) {
  const { t } = useApp();
  return (
    <div className="segmented">
      {items.map((item) => {
        const Icon = item.icon;
        return (
          <button key={item.id} className={value === item.id ? 'active' : ''} onClick={() => onChange(item.id)}>
            <Icon size={16} />
            <span>{t(item.labelKey)}</span>
          </button>
        );
      })}
    </div>
  );
}

function Metric({ label, value, icon: Icon, tone = '', infoKey }) {
  return (
    <div className={`metric ${tone}`}>
      <div>
        <span className="labelWithInfo">{label}{infoKey && <InfoTip textKey={infoKey} />}</span>
        <strong>{value}</strong>
      </div>
      <Icon size={22} />
    </div>
  );
}

function Panel({ title, action, children, infoKey }) {
  return (
    <section className="panel">
      <div className="panelHeader">
        <TitleWithInfo title={title} infoKey={infoKey} />
        {action}
      </div>
      {children}
    </section>
  );
}

function Field({ label, children, infoKey }) {
  return (
    <label className="field">
      <span className="labelWithInfo">{label}{infoKey && <InfoTip textKey={infoKey} />}</span>
      {children}
    </label>
  );
}

function TitleWithInfo({ title, infoKey }) {
  return <h2 className="titleWithInfo">{title}{infoKey && <InfoTip textKey={infoKey} />}</h2>;
}

function InfoTip({ textKey, text }) {
  const { t } = useApp();
  const content = text || t(textKey);
  return (
    <span className="infoTip" tabIndex="0" aria-label={content}>
      <Info size={14} />
      <span className="infoBubble" role="tooltip">{content}</span>
    </span>
  );
}

function ScoreMeter({ label, value, infoKey }) {
  const pct = Math.round((Number(value) || 0) * 100);
  return (
    <div className="scoreMeter">
      <div><span className="labelWithInfo">{label}{infoKey && <InfoTip textKey={infoKey} />}</span><strong>{pct}%</strong></div>
      <div className="meterTrack"><span style={{ width: `${Math.min(100, pct)}%` }} /></div>
    </div>
  );
}

function SignalRow({ signals }) {
  const { t } = useApp();
  if (!signals.length) return null;
  return (
    <div className="signalBlock">
      <span className="labelWithInfo">{t('decision.signals')}<InfoTip textKey="tip.decision.signals" /></span>
      <div className="signalRow">{signals.slice(0, 5).map((signal) => <code key={signal}>{signal}</code>)}</div>
    </div>
  );
}

function ActionList({ title, items, infoKey }) {
  if (!items?.length) return null;
  return (
    <div className="actionList">
      <strong className="labelWithInfo">{title}{infoKey && <InfoTip textKey={infoKey} />}</strong>
      <ul>{items.map((item) => <li key={item}>{item}</li>)}</ul>
    </div>
  );
}

function RiskBadge({ level }) {
  return <span className={`riskBadge ${level || 'LOW'}`}>{level || 'LOW'}</span>;
}

function StatusDot({ ok, label, compact = false }) {
  return <span className={compact ? `statusDot compact ${ok ? 'ok' : 'bad'}` : `statusDot ${ok ? 'ok' : 'bad'}`}><i />{!compact && label}</span>;
}

function Notice({ text }) {
  return <div className="notice"><AlertTriangle size={18} />{text}</div>;
}

function EmptyState({ icon: Icon, text }) {
  return <div className="emptyState"><Icon size={22} /><span>{text}</span></div>;
}

function BucketBars({ items, selected, onSelect }) {
  if (!items.length) return <EmptyState icon={BarChart3} text="No data" />;
  return (
    <div className="bucketBars">
      {items.slice(0, 6).map((item) => (
        <button className={selected === item.label ? 'bucket active' : 'bucket'} key={item.label} onClick={() => onSelect?.(item.label)}>
          <div><span>{item.label}</span><strong>{item.count}</strong></div>
          <div className="meterTrack"><span style={{ width: `${Math.max(4, item.percentage || 0)}%` }} /></div>
        </button>
      ))}
    </div>
  );
}

function DecisionList({ items, selectedId, onSelect, emptyText }) {
  const { t } = useApp();
  if (!items.length) return <EmptyState icon={FileWarning} text={emptyText || t('empty.noDecisions')} />;
  return <div className="decisionList">{items.map((item) => <DecisionListItem key={item.decisionId} item={item} selected={selectedId === item.decisionId} onClick={onSelect ? () => onSelect(item.decisionId) : undefined} />)}</div>;
}

function DecisionListItem({ item, selected, onClick }) {
  const { t } = useApp();
  return (
    <button className={selected ? 'decisionListItem active' : 'decisionListItem'} onClick={onClick} disabled={!onClick}>
      <RiskBadge level={item.riskLevel} />
      <div>
        <strong>{item.scamType}</strong>
        <span>{item.inputType} - {formatTime(item.createdAt)}</span>
        <div className="rowBadges">
          {item.needsHumanReview && <small>{t('analytics.needsReview')}</small>}
          {item.reportId && <small>{t('analytics.hasReport')}</small>}
          {item.payeeHash && <small>{t('analytics.hasPayee')}</small>}
        </div>
      </div>
      <b>{Number(item.score || 0).toFixed(2)}</b>
    </button>
  );
}

function HistoryList({ items, selectedId, onSelect }) {
  const { t } = useApp();
  if (!items.length) return <EmptyState icon={History} text={t('empty.noDecisions')} />;
  return <div className="decisionList compactList">{items.map((item) => <DecisionListItem key={item.decisionId} item={item} selected={selectedId === item.decisionId} onClick={() => onSelect(item.decisionId)} />)}</div>;
}

function MerchantList({ items }) {
  const { t } = useApp();
  if (!items.length) return <EmptyState icon={WalletCards} text={t('empty.noPayees')} />;
  return (
    <div className="merchantList">
      {items.map((item) => (
        <div className="merchantRow" key={item.payeeHash}>
          <div>
            <strong>{Number(item.riskScore).toFixed(2)}</strong>
            <span>{item.complaintCount} {t('table.complaints').toLowerCase()}</span>
          </div>
          <code>{item.payeeHash}</code>
        </div>
      ))}
    </div>
  );
}

function MerchantResult({ item }) {
  return (
    <div className="merchantResult">
      <ScoreMeter label="Risk" value={item.riskScore} />
      <div className="factGrid">
        <span>Complaints <b>{item.complaintCount}</b></span>
        <span>Review <b>{item.needsHumanReview ? 'Yes' : 'No'}</b></span>
      </div>
      <code>{item.payeeHash}</code>
    </div>
  );
}

function ChatPreview({ items }) {
  if (!items.length) return <EmptyState icon={Bot} text="No messages yet." />;
  return (
    <div className="chatPreview">
      {items.map((item, index) => (
        <div key={`${item.from}-${item.createdAt}-${index}`} className={`bubble ${item.from}`}>
          <p>{item.text}</p>
          {item.buttons?.length > 0 && <div className="bubbleButtons">{item.buttons.map((button) => <span key={button}>{button}</span>)}</div>}
        </div>
      ))}
    </div>
  );
}

function Timeline({ items }) {
  if (!items.length) return <EmptyState icon={Activity} text="No events yet." />;
  return (
    <div className="timeline">
      {items.map((item) => (
        <div className="timelineItem" key={item.eventId}>
          <span><Activity size={14} /></span>
          <div>
            <strong>{item.eventType}</strong>
            <p>{item.correlationId || item.causationId || item.producer}</p>
          </div>
          <time>{formatTime(item.createdAt)}</time>
        </div>
      ))}
    </div>
  );
}

function DeveloperDetails({ value }) {
  const { t } = useApp();
  return (
    <details className="developerDetails">
      <summary><Eye size={15} />{t('details.developer')}</summary>
      <pre>{JSON.stringify(value, null, 2)}</pre>
    </details>
  );
}

function IconButton({ label, onClick, icon: Icon }) {
  return (
    <button className="iconButton" onClick={onClick} title={label} aria-label={label}>
      <Icon size={18} />
    </button>
  );
}

function loadLocalHistory(userId) {
  try {
    return JSON.parse(localStorage.getItem(`scamshield.history.${userId}`) || '[]');
  } catch {
    return [];
  }
}

function saveLocalHistory(userId, decision) {
  const next = mergeDecisions([decision], loadLocalHistory(userId)).slice(0, 20);
  localStorage.setItem(`scamshield.history.${userId}`, JSON.stringify(next));
}

function mergeDecisions(primary, secondary) {
  const seen = new Set();
  return [...primary, ...secondary].filter((item) => {
    if (!item?.decisionId || seen.has(item.decisionId)) return false;
    seen.add(item.decisionId);
    return true;
  }).sort((a, b) => new Date(b.createdAt || 0) - new Date(a.createdAt || 0));
}

function mergeConversation(users, replies) {
  return [...users, ...replies].sort((a, b) => new Date(a.createdAt || 0) - new Date(b.createdAt || 0));
}

function matchesDecisionFilters(decision, filters) {
  if (!decision) return false;
  if (filters.scamType && decision.scamType !== filters.scamType) return false;
  if (filters.riskLevel && decision.riskLevel !== filters.riskLevel) return false;
  if (filters.inputType && decision.inputType !== filters.inputType) return false;
  if (filters.language && decision.language !== filters.language) return false;
  if (filters.needsHumanReview === 'true' && !decision.needsHumanReview) return false;
  return true;
}

async function copyToClipboard(text) {
  if (!text) return;
  if (navigator.clipboard?.writeText) {
    await navigator.clipboard.writeText(text);
    return;
  }
  const el = document.createElement('textarea');
  el.value = text;
  document.body.appendChild(el);
  el.select();
  document.execCommand('copy');
  document.body.removeChild(el);
}

function formatTime(value) {
  if (!value) return 'now';
  return new Intl.DateTimeFormat(undefined, { hour: '2-digit', minute: '2-digit' }).format(new Date(value));
}

function formatDateTime(value) {
  if (!value) return 'now';
  return new Intl.DateTimeFormat(undefined, { dateStyle: 'medium', timeStyle: 'short' }).format(new Date(value));
}

createRoot(document.getElementById('root')).render(<App />);
