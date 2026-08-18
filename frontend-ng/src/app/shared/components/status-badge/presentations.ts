export type StatusTone = 'yellow' | 'blue' | 'purple' | 'indigo' | 'teal' | 'green' | 'orange' | 'red' | 'gray' | 'sky';

// Literal class strings so Tailwind generates them .
export const TONE_CLASSES: Record<StatusTone, string> = {
  yellow: 'bg-yellow-100! text-yellow-800! [--t-status:#eab308]',
  blue: 'bg-blue-100! text-blue-800! [--t-status:#3b82f6]',
  purple: 'bg-purple-100! text-purple-800! [--t-status:#a855f7]',
  indigo: 'bg-indigo-100! text-indigo-800! [--t-status:#6366f1]',
  teal: 'bg-teal-100! text-teal-800! [--t-status:#14b8a6]',
  green: 'bg-green-100! text-green-800! [--t-status:#22c55e]',
  orange: 'bg-orange-100! text-orange-800! [--t-status:#f97316]',
  red: 'bg-red-100! text-red-800! [--t-status:#ef4444]',
  gray: 'bg-gray-100! text-gray-800! [--t-status:#6b7280]',
  sky: 'bg-sky-100! text-sky-800! [--t-status:#0ea5e9]',
};

export interface StatusPresentation {
  readonly label: string;
  readonly tone: StatusTone;
}

export type StatusDomain = 'global' | 'league' | 'draft' | 'game' | 'claim';

export const STATUS_PRESENTATIONS: Record<StatusDomain, Record<string, StatusPresentation>> = {
  league: {
    PENDING: { label: 'Pending', tone: 'yellow' },
    SETUP: { label: 'Setup', tone: 'blue' },
    DRAFTING: { label: 'Drafting', tone: 'purple' },
    // eslint-disable-next-line @typescript-eslint/naming-convention -- backend enum string
    POST_DRAFT: { label: 'Post Draft', tone: 'indigo' },
    // eslint-disable-next-line @typescript-eslint/naming-convention -- backend enum string
    TRANSFER_WINDOW: { label: 'Transfer Window', tone: 'teal' },
    // eslint-disable-next-line @typescript-eslint/naming-convention -- backend enum string
    REGULAR_SEASON: { label: 'Regular Season', tone: 'green' },
    // eslint-disable-next-line @typescript-eslint/naming-convention -- backend enum string
    POST_REGULAR_SEASON: { label: 'Post Regular Season', tone: 'sky' },
    PLAYOFFS: { label: 'Playoffs', tone: 'orange' },
    COMPLETED: { label: 'Completed', tone: 'gray' },
    CANCELLED: { label: 'Cancelled', tone: 'red' },
  },
  draft: {
    PENDING: { label: 'Pending', tone: 'yellow' },
    ONGOING: { label: 'Ongoing', tone: 'purple' },
    PAUSED: { label: 'Paused', tone: 'orange' },
    COMPLETED: { label: 'Completed', tone: 'gray' },
  },
  game: {
    SCHEDULED: { label: 'Scheduled', tone: 'blue' },
    // eslint-disable-next-line @typescript-eslint/naming-convention -- backend enum string
    APPROVAL_PENDING: { label: 'Approval Pending', tone: 'yellow' },
    COMPLETED: { label: 'Completed', tone: 'green' },
    DISPUTED: { label: 'Disputed', tone: 'red' },
  },
  claim: {
    DRAFT: { label: 'Draft', tone: 'purple' },
    // eslint-disable-next-line @typescript-eslint/naming-convention -- backend enum string
    FREE_AGENT: { label: 'Free Agent', tone: 'blue' },
  },
  global: {},
};
