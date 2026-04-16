import { z } from 'zod'

export const hostEventSchema = z.object({
  id: z.string(),
  slug: z.string(),
  title: z.string(),
  event_type: z.string().optional(),
  date_start: z.string().optional(),
})

export const hostEventDetailSchema = hostEventSchema.extend({
  couple_name_a: z.unknown().optional(),
  couple_name_b: z.unknown().optional(),
  date_start: z.unknown().optional(),
  date_end: z.unknown().optional(),
})

export const hostEventsResponseSchema = z.object({
  events: z.array(hostEventSchema).default([]),
})

export const hostShagunItemSchema = z.object({
  id: z.string(),
  channel: z.string(),
  amount_paise: z.number().optional().default(0),
  status: z.string(),
  created_at: z.string(),
  meta: z.object({ sender_name: z.string().optional() }).optional(),
})

export const hostShagunResponseSchema = z.object({
  shagun: z.array(hostShagunItemSchema).default([]),
})

export const hostGuestSchema = z.object({
  id: z.string(),
  name: z.string(),
  phone: z.string(),
  priority_score: z.number().optional().default(0),
  priority_tier: z.string().optional().default('Optional'),
  priority_reasons: z.array(z.string()).optional().default([]),
})

export const hostGuestsResponseSchema = z.object({
  guests: z.array(hostGuestSchema).default([]),
  limit: z.number().optional().default(50),
  offset: z.number().optional().default(0),
  sort: z.string().optional().default('name_asc'),
  priority_tier: z.string().optional().default(''),
  /** Opaque token for the next page when using keyset (cursor) pagination. */
  next_cursor: z.string().optional(),
})

export const hostGuestsImportResponseSchema = z.object({
  imported: z.number(),
  errors: z.array(z.object({ line: z.number(), error: z.string() })).default([]),
})

export const hostRelationshipPriorityOverviewSchema = z.object({
  feature: z.string(),
  status: z.string(),
  ranked_guests: z.array(hostGuestSchema).default([]),
  guests_needing_attention: z.array(hostGuestSchema).default([]),
  tier_counts: z.object({
    critical: z.number().optional().default(0),
    important: z.number().optional().default(0),
    optional: z.number().optional().default(0),
  }),
  coming_next: z.array(z.string()).default([]),
})

export const hostSubEventSchema = z.object({
  id: z.string(),
  name: z.string(),
})

export const hostSubEventsResponseSchema = z.object({
  sub_events: z.array(hostSubEventSchema).default([]),
})

export const hostVendorSchema = z.object({
  id: z.string(),
  name: z.string(),
  category: z.string().optional().default(''),
  phone: z.string().optional().default(''),
  email: z.string().optional().default(''),
  advance_paise: z.number().optional().default(0),
  total_paise: z.number().optional().default(0),
  notes: z.string().optional().default(''),
  status: z.string().optional().default(''),
})

export const hostVendorsResponseSchema = z.object({
  vendors: z.array(hostVendorSchema).default([]),
})

export const hostBroadcastSchema = z.object({
  id: z.string(),
  title: z.string(),
  body: z.string(),
  announcement_type: z.string(),
  audience: z.string(),
  created_at: z.string(),
})

export const hostBroadcastsResponseSchema = z.object({
  broadcasts: z.array(hostBroadcastSchema).default([]),
})

export const hostRSVPRowSchema = z.object({
  id: z.string(),
  guest_phone: z.string(),
  sub_event_id: z.string(),
  status: z.string(),
})

export const hostRSVPResponseSchema = z.object({
  rsvps: z.array(hostRSVPRowSchema).default([]),
})

export const hostGalleryAssetSchema = z.object({
  id: z.string(),
  section: z.string(),
  object_key: z.string(),
  status: z.enum(['pending', 'approved', 'rejected']),
  url: z.string().optional(),
  created_at: z.string().optional(),
})

export const hostGalleryAssetsResponseSchema = z.object({
  assets: z.array(hostGalleryAssetSchema).default([]),
})

export const hostGalleryPresignResponseSchema = z.object({
  upload: z.object({
    method: z.string(),
    url: z.string(),
    headers: z.record(z.string(), z.string()),
    object_key: z.string(),
  }),
})

export const hostBillingCheckoutSchema = z.object({
  id: z.string(),
  tier: z.string(),
  status: z.string(),
  order_id: z.string(),
  event_id: z.string().optional(),
})

export const hostBillingCheckoutsResponseSchema = z.object({
  checkouts: z.array(hostBillingCheckoutSchema).default([]),
})

export const organiserClientSchema = z.object({
  id: z.string(),
  name: z.string(),
  contact_email: z.string().optional().nullable(),
  contact_phone: z.string().optional().nullable(),
  notes: z.string().optional().nullable(),
})

export const organiserClientsResponseSchema = z.object({
  clients: z.array(organiserClientSchema).default([]),
})

export const hostMemoryBookGenerateResponseSchema = z.object({
  slug: z.string(),
  public_api_path: z.string(),
  payload: z.record(z.string(), z.unknown()),
  export_pdf_available: z.boolean(),
})

export const hostMemoryBookExportResponseSchema = z.object({
  status: z.string().optional(),
  next_step: z.string().optional(),
})

export function parseHostEvent(input: unknown) {
  return hostEventSchema.parse(input)
}

export function parseHostEventDetail(input: unknown) {
  return hostEventDetailSchema.parse(input)
}

export function parseHostEventsResponse(input: unknown) {
  return hostEventsResponseSchema.parse(input)
}

export function parseHostShagunResponse(input: unknown) {
  return hostShagunResponseSchema.parse(input)
}

export function parseHostGuestsResponse(input: unknown) {
  return hostGuestsResponseSchema.parse(input)
}

export function parseHostGuestsImportResponse(input: unknown) {
  return hostGuestsImportResponseSchema.parse(input)
}

export function parseHostRelationshipPriorityOverview(input: unknown) {
  return hostRelationshipPriorityOverviewSchema.parse(input)
}

export function parseHostSubEventsResponse(input: unknown) {
  return hostSubEventsResponseSchema.parse(input)
}

export function parseHostVendorsResponse(input: unknown) {
  return hostVendorsResponseSchema.parse(input)
}

export function parseHostBroadcastsResponse(input: unknown) {
  return hostBroadcastsResponseSchema.parse(input)
}

export function parseHostRSVPResponse(input: unknown) {
  return hostRSVPResponseSchema.parse(input)
}

export function parseHostGalleryAssetsResponse(input: unknown) {
  return hostGalleryAssetsResponseSchema.parse(input)
}

export function parseHostGalleryPresignResponse(input: unknown) {
  return hostGalleryPresignResponseSchema.parse(input)
}

export function parseHostBillingCheckoutsResponse(input: unknown) {
  return hostBillingCheckoutsResponseSchema.parse(input)
}

export function parseOrganiserClientsResponse(input: unknown) {
  return organiserClientsResponseSchema.parse(input)
}

export function parseHostMemoryBookGenerateResponse(input: unknown) {
  return hostMemoryBookGenerateResponseSchema.parse(input)
}

export function parseHostMemoryBookExportResponse(input: unknown) {
  return hostMemoryBookExportResponseSchema.parse(input)
}
