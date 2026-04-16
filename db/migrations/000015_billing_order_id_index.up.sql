CREATE UNIQUE INDEX IF NOT EXISTS idx_billing_checkouts_razorpay_order_id
    ON billing_checkouts (razorpay_order_id)
    WHERE razorpay_order_id IS NOT NULL;
