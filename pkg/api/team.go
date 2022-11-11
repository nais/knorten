package api

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/nais/knorten/pkg/chart"
	"github.com/nais/knorten/pkg/database/gensql"
	"net/http"
)

func (a *API) setupTeamRoutes() {
	a.router.GET("/team/new", func(c *gin.Context) {
		var form chart.NamespaceForm
		// err := c.ShouldBind(&form)
		c.HTML(http.StatusOK, "charts/namespace.tmpl", gin.H{
			"form": form,
		})
	})

	a.router.POST("/team/new", func(c *gin.Context) {
		err := chart.CreateNamespace(c, a.repo, a.helmClient, gensql.ChartTypeNamespace, a.dryRun)
		if err != nil {
			fmt.Println(err)
			// c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			c.Redirect(http.StatusSeeOther, "/team/new")
		}
		c.Redirect(http.StatusSeeOther, "/user")
	})

	a.router.GET("/team/:team/edit", func(c *gin.Context) {
		team := c.Param("team")
		namespaceForm := &chart.NamespaceForm{}
		err := a.repo.TeamConfigurableValuesGet(c, gensql.ChartTypeNamespace, team, namespaceForm)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.HTML(http.StatusOK, "charts/namespace.tmpl", gin.H{
			"values": namespaceForm,
			"team":   team,
		})
	})

	a.router.POST("/team/:team/edit", func(c *gin.Context) {
		err := chart.UpdateNamespace(c, a.helmClient, a.repo)
		if err != nil {
			fmt.Println(err)
			team := c.Param("team")
			// c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			c.Redirect(http.StatusSeeOther, fmt.Sprintf("/team/%v/new", team))
			return
		}
		c.Redirect(http.StatusSeeOther, "/user")
	})
}
