package router

import (
	"net/http"

	controller "core/api/controller"
	config "core/internal/config"
	service "core/internal/service"
	gmail "core/pkg/gmail"
	openai "core/pkg/openai"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func Config(db *gorm.DB, cfg *config.Config) http.Handler {
	router := gin.Default()
	router.MaxMultipartMemory = 8 << 20

	controllers := controller.Create(
		db,
		service.NewResearch(db, openai.Client),
		service.NewGmail(gmail.Client, &cfg.Gmail),
	)

	router.Use(cors.New(cors.Config{
		AllowCredentials: true,
		AllowHeaders:     []string{"Content-Type", "Accept"},
		AllowOrigins:     []string{cfg.Server.AppOrigin},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE"},
	}))

	v1 := router.Group("/api/v1")
	{
		v1.GET("/health", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "healthy"})
		})

		v1.GET("/integrations", controllers.GetIntegrations)
		v1.POST("/integrations/gmail/sync", controllers.SyncGmail)

		campaigns := v1.Group("/campaigns")
		{
			campaigns.GET("", controllers.GetCampaigns)
			campaigns.POST("", controllers.CreateCampaign)
			campaigns.GET("/:id", controllers.GetCampaign)
			campaigns.PATCH("/:id", controllers.UpdateCampaign)
			campaigns.DELETE("/:id", controllers.DeleteCampaign)
			campaigns.GET("/:id/stats", controllers.GetCampaignStats)
			campaigns.GET("/:id/icp", controllers.GetICP)
			campaigns.PUT("/:id/icp", controllers.UpdateICP)
			campaigns.POST("/:id/companies/import", controllers.ImportCompanies)
		}

		companies := v1.Group("/companies")
		{
			companies.GET("", controllers.GetCompanies)
			companies.POST("", controllers.CreateCompany)
			companies.GET("/:id", controllers.GetCompany)
			companies.PATCH("/:id", controllers.UpdateCompany)
			companies.DELETE("/:id", controllers.DeleteCompany)
			companies.POST("/:id/findings", controllers.CreateFinding)
			companies.DELETE("/:id/findings/:findingId", controllers.DeleteFinding)
			companies.POST("/:id/research", controllers.RunResearch)
			companies.GET("/:id/messages", controllers.GetMessages)
			companies.POST("/:id/messages", controllers.CreateMessage)
		}

		inbound := v1.Group("/inbound")
		{
			inbound.GET("", controllers.GetOpportunities)
			inbound.POST("", controllers.CreateOpportunity)
			inbound.GET("/:id", controllers.GetOpportunity)
			inbound.PATCH("/:id", controllers.UpdateOpportunity)
		}

		drafts := v1.Group("/drafts")
		{
			drafts.GET("", controllers.GetDrafts)
			drafts.POST("", controllers.CreateDraft)
			drafts.PATCH("/:id", controllers.UpdateDraft)
			drafts.DELETE("/:id", controllers.DeleteDraft)
		}
	}

	return router
}
