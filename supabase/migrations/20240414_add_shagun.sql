-- UTSAV Migration: Digital Shagun & Financials

-- 1. Add UPI ID and Branding Color to events table
ALTER TABLE events ADD COLUMN IF NOT EXISTS upi_id TEXT;
ALTER TABLE events ADD COLUMN IF NOT EXISTS branding_color TEXT;

-- 2. Create Shagun table
CREATE TABLE IF NOT EXISTS shagun (
  id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
  event_id UUID REFERENCES events(id) ON DELETE CASCADE NOT NULL,
  sender_id UUID REFERENCES profiles(id), -- Optional
  sender_name TEXT NOT NULL,
  amount DECIMAL(10, 2) NOT NULL,
  message TEXT,
  payment_method TEXT DEFAULT 'UPI', -- UPI, CASH, GIFT
  status TEXT DEFAULT 'GUEST_REPORTED', -- GUEST_REPORTED, VERIFIED, DISPUTED
  transaction_id TEXT,
  created_at TIMESTAMPTZ DEFAULT NOW()
);

-- 3. Enable RLS
ALTER TABLE shagun ENABLE ROW LEVEL SECURITY;

-- 4. RLS Policies
CREATE POLICY "Anyone can record shagun for an event" ON shagun 
  FOR INSERT WITH CHECK (true);

CREATE POLICY "Owners can view shagun for their events" ON shagun 
  FOR SELECT USING (
    EXISTS (
      SELECT 1 FROM events 
      WHERE events.id = shagun.event_id 
      AND events.owner_id = auth.uid()
    )
  );

CREATE POLICY "Owners can update shagun status" ON shagun 
  FOR UPDATE USING (
    EXISTS (
      SELECT 1 FROM events 
      WHERE events.id = shagun.event_id 
      AND events.owner_id = auth.uid()
    )
  );
