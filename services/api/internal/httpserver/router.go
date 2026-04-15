package httpserver

import (
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func (s *Server) Mount(r *gin.Engine) {
	r.GET("/health", s.healthz)
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))
	v1 := r.Group("/v1")
	v1.GET("/healthz", s.healthz)
	v1.GET("/readyz", s.readyz)

	v1.POST("/auth/otp/request", s.postOTPRequest)
	v1.POST("/auth/otp/verify", s.postOTPVerify)
	v1.POST("/auth/refresh", s.postRefresh)
	v1.POST("/auth/logout", s.postLogout)

	authed := v1.Group("/")
	authed.Use(s.requireUserMiddleware())
	authed.GET("/me", s.getMe)

	authed.GET("/events/check-slug", s.getCheckSlug)
	authed.POST("/events", s.postEvent)
	authed.GET("/events", s.listEvents)

	eventAuthed := authed.Group("/events/:id")
	eventAuthed.Use(s.requireEventAccessMiddleware())
	eventAuthed.GET("", s.getEvent)
	eventAuthed.PATCH("", s.patchEvent)
	eventAuthed.POST("/sub-events", s.postSubEvent)
	eventAuthed.GET("/sub-events", s.listSubEvents)
	eventAuthed.POST("/members", s.postEventMember)

	eventAuthed.GET("/guests", s.listGuests)
	eventAuthed.POST("/guests", s.postGuest)
	eventAuthed.POST("/guests/import", s.postGuestsImport)
	eventAuthed.GET("/vendors", s.listVendors)
	eventAuthed.POST("/vendors", s.postVendor)
	eventAuthed.DELETE("/vendors/:vendorId", s.deleteVendor)
	eventAuthed.POST("/cash-shagun", s.postCashShagun)

	eventAuthed.GET("/rsvps", s.listRSVPsHost)
	eventAuthed.GET("/shagun", s.listShagunHost)

	eventAuthed.POST("/gallery/assets", s.postGalleryAsset)
	eventAuthed.POST("/gallery/presign", s.postGalleryPresign)
	eventAuthed.GET("/gallery/assets", s.listGalleryAssets)
	eventAuthed.PATCH("/gallery/assets/:assetId", s.patchGalleryAssetModeration)
	eventAuthed.GET("/broadcasts", s.listBroadcastsHost)
	eventAuthed.POST("/broadcasts", s.postBroadcast)
	eventAuthed.POST("/memory-book/generate", s.postMemoryBookGenerate)
	eventAuthed.POST("/memory-book/export", s.postMemoryBookExport)

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

	authed.POST("/organiser/profile", s.postOrganiserProfile)
	authed.GET("/organiser/me", s.getOrganiserMe)
	authed.GET("/organiser/events", s.getOrganiserEvents)
	authed.GET("/organiser/clients", s.listOrganiserClients)
	authed.POST("/organiser/clients", s.postOrganiserClient)
	authed.PATCH("/organiser/clients/:clientId", s.patchOrganiserClient)
	authed.POST("/organiser/clients/:clientId/events", s.postOrganiserClientEvent)

	authed.POST("/billing/checkout", s.postBillingCheckout)
	authed.GET("/billing/checkouts", s.listBillingCheckouts)
	v1.POST("/billing/webhook/razorpay", s.postRazorpayWebhook)
}
