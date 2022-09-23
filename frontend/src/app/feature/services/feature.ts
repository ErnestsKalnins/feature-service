export interface Feature {
  id: string | null,
  displayName: string | null,
  technicalName: string,
  description: string | null,
  expiresOn: string | null,
  inverted: boolean,
}
