export interface Feature {
  id: string | null,
  displayName: string | null,
  technicalName: string,
  description: string | null,
  expiresOn: number | null,
  inverted: boolean,
  createdAt: number,
  updatedAt: number,
  customerIds: string[] | null,
}
