import { z } from 'zod'

export const hostEventSchema = z.object({
  id: z.string(),
  slug: z.string(),
  title: z.string(),
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
})

export const hostGuestsResponseSchema = z.object({
  guests: z.array(hostGuestSchema).default([]),
})

export const hostGuestsImportResponseSchema = z.object({
  imported: z.number(),
  errors: z.array(z.object({ line: z.number(), error: z.string() })).default([]),
})

export const hostSubEventSchema = z.object({
  id: z.string(),
  name: z.string(),
})

export const hostSubEventsResponseSchema = z.object({
  sub_events: z.array(hostSubEventSchema).default([]),
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

export function parseHostEvent(input: unknown) {
  return hostEventSchema.parse(input)
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

export function parseHostSubEventsResponse(input: unknown) {
  return hostSubEventsResponseSchema.parse(input)
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
