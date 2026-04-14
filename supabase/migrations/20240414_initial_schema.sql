-- UTSAV Database Schema
-- Phase 1 Foundation & Auth

-- 1. Enums
CREATE TYPE event_role AS ENUM ('OWNER', 'CO_OWNER', 'ORGANISER', 'CONTRIBUTOR', 'VENDOR', 'GUEST');
CREATE TYPE event_type AS ENUM ('WEDDING', 'BIRTHDAY', 'PARTY', 'GET_TOGETHER');
CREATE TYPE rsvp_status AS ENUM ('PENDING', 'CONFIRMED', 'DECLINED', 'MAYBE');

-- 2. Profiles (Extends Auth.Users)
CREATE TABLE profiles (
  id UUID REFERENCES auth.users ON DELETE CASCADE PRIMARY KEY,
  phone TEXT UNIQUE NOT NULL,
  full_name TEXT,
  avatar_url TEXT,
  created_at TIMESTAMPTZ DEFAULT NOW(),
  updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- 3. Events
CREATE TABLE events (
  id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
  title TEXT NOT NULL,
  slug TEXT UNIQUE NOT NULL,
  type event_type NOT NULL DEFAULT 'WEDDING',
  description TEXT,
  cover_image TEXT,
  owner_id UUID REFERENCES profiles(id) ON DELETE CASCADE NOT NULL,
  start_date DATE NOT NULL,
  end_date DATE NOT NULL,
  is_public BOOLEAN DEFAULT false,
  settings JSONB DEFAULT '{
    "shagun_enabled": true,
    "gallery_enabled": true,
    "rsvp_enabled": true
  }'::JSONB,
  branding JSONB DEFAULT '{
    "primary_color": "#EA580C",
    "theme_name": "default"
  }'::JSONB,
  created_at TIMESTAMPTZ DEFAULT NOW()
);

-- 4. Sub-events (Schedule items)
CREATE TABLE sub_events (
  id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
  event_id UUID REFERENCES events(id) ON DELETE CASCADE NOT NULL,
  name TEXT NOT NULL,
  type TEXT, -- e.g. "Sangeet", "Haldi"
  date_time TIMESTAMPTZ NOT NULL,
  venue_name TEXT,
  venue_address TEXT,
  dress_code TEXT,
  description TEXT,
  created_at TIMESTAMPTZ DEFAULT NOW()
);

-- 5. Event Members (Admin/Staff Access)
CREATE TABLE event_members (
  id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
  event_id UUID REFERENCES events(id) ON DELETE CASCADE NOT NULL,
  user_id UUID REFERENCES profiles(id) ON DELETE CASCADE NOT NULL,
  role event_role NOT NULL DEFAULT 'GUEST',
  invited_by UUID REFERENCES profiles(id),
  created_at TIMESTAMPTZ DEFAULT NOW(),
  UNIQUE(event_id, user_id)
);

-- 6. Guest List (Tracking all attendees)
CREATE TABLE guests (
  id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
  event_id UUID REFERENCES events(id) ON DELETE CASCADE NOT NULL,
  name TEXT NOT NULL,
  phone TEXT,
  email TEXT,
  relationship TEXT, -- college friends, family, etc.
  side TEXT, -- bride/groom
  group_id TEXT, -- for family RSVP
  status rsvp_status DEFAULT 'PENDING',
  notes TEXT,
  created_at TIMESTAMPTZ DEFAULT NOW()
);

-- RLS POLICIES (Simplified for MVP)

-- Profiles: Public can see profiles (to show names), owners can edit
ALTER TABLE profiles ENABLE ROW LEVEL SECURITY;
CREATE POLICY "Public profiles are viewable by everyone" ON profiles FOR SELECT USING (true);
CREATE POLICY "Users can update own profile" ON profiles FOR UPDATE USING (auth.uid() = id);

-- Events: Members can view. Owner can manage.
ALTER TABLE events ENABLE ROW LEVEL SECURITY;
CREATE POLICY "Public events are viewable by everyone" ON events FOR SELECT USING (is_public = true);
CREATE POLICY "Members can view private events" ON events FOR SELECT USING (
  EXISTS (SELECT 1 FROM event_members WHERE event_id = id AND user_id = auth.uid())
);
CREATE POLICY "Owners can manage events" ON events ALL USING (owner_id = auth.uid());

-- Sub-events: Inherit from event view permissions
ALTER TABLE sub_events ENABLE ROW LEVEL SECURITY;
CREATE POLICY "Viewable by event viewers" ON sub_events FOR SELECT USING (
  EXISTS (SELECT 1 FROM events WHERE id = event_id AND (is_public = true OR EXISTS (SELECT 1 FROM event_members WHERE event_id = events.id AND user_id = auth.uid())))
);
