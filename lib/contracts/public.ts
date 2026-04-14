import { z } from 'zod'

const publicProfileSchema = z.object({
  full_name: z.string().optional(),
})

const publicSubEventSchema = z.object({
  id: z.string(),
  name: z.string(),
  type: z.string().optional(),
  sub_type: z.string().optional(),
  date_time: z.string().optional(),
  starts_at: z.string().optional(),
  venue_name: z.string().optional(),
  venue_label: z.string().optional(),
  venue_address: z.string().optional(),
  dress_code: z.string().optional(),
})

export const publicEventSchema = z.object({
  id: z.string(),
  slug: z.string().optional(),
  title: z.string(),
  event_type: z.string().optional(),
  description: z.string().optional(),
  branding_color: z.string().optional(),
  cover_image: z.string().optional(),
  cover_image_url: z.string().optional(),
  start_date: z.string().optional(),
  date_start: z.string().optional(),
  host_upi_vpa: z.string().optional(),
  upi_id: z.string().optional(),
  profiles: publicProfileSchema.optional(),
  sub_events: z.array(publicSubEventSchema).optional(),
})

export const publicEventResponseSchema = z.object({
  event: publicEventSchema,
})

export const publicGalleryResponseSchema = z.object({
  assets: z.array(z.object({ id: z.string() })).optional(),
})

export const memoryPayloadSchema = z.object({
  payload: z
    .object({
      highlights: z
        .object({
          shagun_count: z.number().optional(),
          shagun_total_paise: z.number().optional(),
        })
        .optional(),
      featured_wishes: z
        .array(
          z.object({
            id: z.string(),
            blessing_note: z.string().optional(),
            meta: z.object({ sender_name: z.string().optional() }).optional(),
          })
        )
        .optional(),
    })
    .optional(),
})

export function parsePublicEventResponse(input: unknown) {
  return publicEventResponseSchema.parse(input)
}

export function parsePublicGalleryResponse(input: unknown) {
  return publicGalleryResponseSchema.parse(input)
}

export function parseMemoryPayloadResponse(input: unknown) {
  return memoryPayloadSchema.parse(input)
}
