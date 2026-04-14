export type Role = 'OWNER' | 'CO_OWNER' | 'ORGANISER' | 'CONTRIBUTOR' | 'VENDOR' | 'GUEST';

export interface Profile {
  id: string;
  phone: string;
  full_name: string | null;
  avatar_url: string | null;
  created_at: string;
}

export interface Event {
  id: string;
  title: string;
  slug: string;
  event_type: 'WEDDING' | 'BIRTHDAY' | 'PARTY' | 'GET_TOGETHER';
  description: string | null;
  cover_image: string | null;
  owner_id: string;
  start_date: string;
  end_date: string;
  is_public: boolean;
  settings: {
    shagun_enabled: boolean;
    gallery_enabled: boolean;
    rsvp_enabled: boolean;
  };
  branding: {
    theme_name: string;
    primary_color?: string;
  };
  upi_id?: string;
  created_at: string;
}

export interface SubEvent {
  id: string;
  event_id: string;
  name: string;
  type: string;
  date_time: string;
  venue_name: string;
  venue_address: string | null;
  dress_code: string | null;
  description: string | null;
}

export interface RoleAssignment {
  id: string;
  event_id: string;
  user_phone: string;
  role: Role;
  assigned_at: string;
}
