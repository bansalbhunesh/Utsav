package httpserver

import "github.com/gin-gonic/gin"

func (s *Server) Mount(r *gin.Engine) {
	v1 := r.Group("/v1")
	v1.GET("/healthz", s.healthz)
	v1.GET("/readyz", s.readyz)

	v1.POST("/auth/otp/request", s.postOTPRequest)
	v1.POST("/auth/otp/verify", s.postOTPVerify)
	v1.POST("/auth/refresh", s.postRefresh)

	v1.GET("/me", s.getMe)

	v1.GET("/events/check-slug", s.getCheckSlug)
	v1.POST("/events", s.postEvent)
	v1.GET("/events", s.listEvents)
	v1.GET("/events/:id", s.getEvent)
	v1.PATCH("/events/:id", s.patchEvent)
	v1.POST("/events/:id/sub-events", s.postSubEvent)
	v1.GET("/events/:id/sub-events", s.listSubEvents)
	v1.POST("/events/:id/members", s.postEventMember)

	v1.GET("/events/:id/guests", s.listGuests)
	v1.POST("/events/:id/guests", s.postGuest)
	v1.POST("/events/:id/guests/import", s.postGuestsImport)
	v1.POST("/events/:id/cash-shagun", s.postCashShagun)

	v1.GET("/events/:id/rsvps", s.listRSVPsHost)
	v1.GET("/events/:id/shagun", s.listShagunHost)

	v1.POST("/events/:id/gallery/assets", s.postGalleryAsset)
	v1.POST("/events/:id/gallery/presign", s.postGalleryPresign)
	v1.GET("/events/:id/gallery/assets", s.listGalleryAssets)
	v1.PATCH("/events/:id/gallery/assets/:assetId", s.patchGalleryAssetModeration)
	v1.GET("/events/:id/broadcasts", s.listBroadcastsHost)
	v1.POST("/events/:id/broadcasts", s.postBroadcast)
	v1.POST("/events/:id/memory-book/generate", s.postMemoryBookGenerate)
	v1.POST("/events/:id/memory-book/export", s.postMemoryBookExport)

	v1.GET("/public/events/:slug", s.getPublicEvent)
	v1.GET("/public/events/:slug/schedule", s.getPublicSchedule)
	v1.GET("/public/events/:slug/gallery", s.getPublicGallery)
	v1.GET("/public/events/:slug/broadcasts", s.getPublicBroadcasts)
	v1.GET("/public/events/:slug/upi-link", s.getPublicUPILink)

	v1.POST("/public/events/:slug/rsvp/otp/request", s.postPublicRSVPOTPRequest)
	v1.POST("/public/events/:slug/rsvp/otp/verify", s.postPublicRSVPOTPVerify)
	v1.POST("/public/events/:slug/rsvp", s.postPublicRSVP)
	v1.POST("/public/events/:slug/shagun/report", s.postPublicShagunReport)

	v1.GET("/public/memory/:slug", s.getPublicMemoryBook)

	v1.POST("/organiser/profile", s.postOrganiserProfile)
	v1.GET("/organiser/me", s.getOrganiserMe)
	v1.GET("/organiser/events", s.getOrganiserEvents)
	v1.GET("/organiser/clients", s.listOrganiserClients)
	v1.POST("/organiser/clients", s.postOrganiserClient)
	v1.PATCH("/organiser/clients/:clientId", s.patchOrganiserClient)
	v1.POST("/organiser/clients/:clientId/events", s.postOrganiserClientEvent)

	v1.POST("/billing/checkout", s.postBillingCheckout)
	v1.GET("/billing/checkouts", s.listBillingCheckouts)
	v1.POST("/billing/webhook/razorpay", s.postRazorpayWebhook)
}
