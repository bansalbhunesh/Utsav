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
