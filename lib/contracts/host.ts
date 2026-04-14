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

export function parseHostEvent(input: unknown) {
  return hostEventSchema.parse(input)
}

export function parseHostShagunResponse(input: unknown) {
  return hostShagunResponseSchema.parse(input)
}
