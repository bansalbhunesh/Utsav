export interface PublicProfile {
  full_name?: string
}

export interface PublicSubEvent {
  id: string
  name: string
  type?: string
  sub_type?: string
  date_time?: string
  starts_at?: string
  venue_name?: string
  venue_label?: string
  venue_address?: string
  dress_code?: string
}

export interface PublicEvent {
  id: string
  slug?: string
  title: string
  event_type?: string
  description?: string
  branding_color?: string
  cover_image?: string
  cover_image_url?: string
  start_date?: string
  date_start?: string
  host_upi_vpa?: string
  upi_id?: string
  profiles?: PublicProfile
  sub_events?: PublicSubEvent[]
}

export interface PublicEventResponse {
  event: PublicEvent
}
