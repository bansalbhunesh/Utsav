-- UTSAV Migration: Gallery & Logistics

-- 1. Create Media Table (Gallery)
CREATE TABLE IF NOT EXISTS event_media (
  id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
  event_id UUID REFERENCES events(id) ON DELETE CASCADE NOT NULL,
  user_id UUID REFERENCES profiles(id),
  url TEXT NOT NULL,
  type TEXT DEFAULT 'IMAGE', -- IMAGE, VIDEO
  is_official BOOLEAN DEFAULT false,
  caption TEXT,
  created_at TIMESTAMPTZ DEFAULT NOW()
);

-- 2. Create Vendors Table
CREATE TABLE IF NOT EXISTS vendors (
  id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
  event_id UUID REFERENCES events(id) ON DELETE CASCADE NOT NULL,
  name TEXT NOT NULL,
  category TEXT, -- Catering, Decoration, Photography, etc.
  contact_info TEXT,
  budget_amount DECIMAL(10, 2),
  paid_amount DECIMAL(10, 2) DEFAULT 0,
  status TEXT DEFAULT 'PENDING', -- PENDING, HIRED, COMPLETED
  notes TEXT,
  created_at TIMESTAMPTZ DEFAULT NOW()
);

-- 3. Enable RLS
ALTER TABLE event_media ENABLE ROW LEVEL SECURITY;
ALTER TABLE vendors ENABLE ROW LEVEL SECURITY;

-- 4. RLS Policies
CREATE POLICY "Anyone can view media for public events" ON event_media
  FOR SELECT USING (
    EXISTS (SELECT 1 FROM events WHERE id = event_id AND is_public = true)
  );

CREATE POLICY "Owners can manage media" ON event_media
  ALL USING (
    EXISTS (SELECT 1 FROM events WHERE id = event_id AND owner_id = auth.uid())
  );

CREATE POLICY "Owners can manage vendors" ON vendors
  ALL USING (
    EXISTS (SELECT 1 FROM events WHERE id = event_id AND owner_id = auth.uid())
  );
