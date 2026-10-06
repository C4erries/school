export interface CalendarSettings {
  calendar_token: string;
  feed_url: string;
  webcal_url: string;
}

export type CalendarSettingsResponse = CalendarSettings;

export interface CalendarExportParams {
  from?: string;
  to?: string;
}
