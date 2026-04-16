export interface PaymentDetails {
  amount: number;
  receiverUpiId: string;
  receiverName: string;
  transactionNote?: string;
}

/**
 * PaymentService handles routing to UPI payment apps using deep linking.
 * This implementation is for Web PWA, using window.location for intents.
 */
class PaymentService {
  /**
   * Generates a UPI intent URL and redirects the user to their UPI app.
   */
  async processPayment(details: PaymentDetails): Promise<boolean> {
    const { amount, receiverUpiId, receiverName, transactionNote } = details;
    
    // The strict UPI deep-link format
    // pa = Payee Address (VPA)
    // pn = Payee Name
    // am = Amount
    // cu = Currency (INR)
    // tn = Transaction Note
    const upiUrl = `upi://pay?pa=${receiverUpiId}&pn=${encodeURIComponent(receiverName)}&am=${amount}&cu=INR${transactionNote ? `&tn=${encodeURIComponent(transactionNote)}` : ''}`;

    try {
      // In a Web context, we just try to open the URL.
      // If no UPI app is installed, the browser might show an error or do nothing.
      window.location.href = upiUrl;
      
      // Since we can't get a callback from the UPI intent on Web, 
      // the caller should show a confirmation dialog to the user.
      return true;
    } catch (e) {
      console.error('Failed to initiate UPI payment', e);
      return false;
    }
  }

  /**
   * Formats a number as INR (₹)
   */
  formatINR(amount: number): string {
    return new Intl.NumberFormat('en-IN', {
      style: 'currency',
      currency: 'INR',
      maximumFractionDigits: 0,
    }).format(amount);
  }
}

export const paymentService = new PaymentService();
