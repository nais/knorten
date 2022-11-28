package api

import (
	"fmt"
	"net/http"

	"github.com/go-playground/validator/v10"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/nais/knorten/pkg/chart"
	"github.com/nais/knorten/pkg/database/gensql"
)

func getChartType(chartType string) gensql.ChartType {
	switch chartType {
	case string(gensql.ChartTypeJupyterhub):
		return gensql.ChartTypeJupyterhub
	case string(gensql.ChartTypeAirflow):
		return gensql.ChartTypeAirflow
	default:
		return ""
	}
}

func (a *API) setupChartRoutes() {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		err := v.RegisterValidation("validDagRepo", chart.AirflowValidateDagRepo)
		if err != nil {
			a.log.WithError(err).Error("can't register validator")
			return
		}
	}

	a.router.GET("/team/:team/:chart/new", func(c *gin.Context) {
		team := c.Param("team")
		chartType := getChartType(c.Param("chart"))

		var form any

		switch chartType {
		case gensql.ChartTypeJupyterhub:
			form = chart.JupyterForm{}
		case gensql.ChartTypeAirflow:
			form = chart.AirflowForm{}
		}

		session := sessions.Default(c)
		flashes := session.Flashes()
		err := session.Save()
		if err != nil {
			a.log.WithError(err).Error("problem saving session")
			return
		}

		c.HTML(http.StatusOK, fmt.Sprintf("charts/%v", chartType), gin.H{
			"team":   team,
			"form":   form,
			"errors": flashes,
		})
	})

	a.router.POST("/team/:team/:chart/new", func(c *gin.Context) {
		slug := c.Param("team")
		chartType := getChartType(c.Param("chart"))
		var err error

		switch chartType {
		case gensql.ChartTypeJupyterhub:
			err = chart.CreateJupyterhub(c, slug, a.repo, a.helmClient, a.cryptor)
		case gensql.ChartTypeAirflow:
			err = chart.CreateAirflow(c, slug, a.repo, a.googleClient, a.k8sClient, a.helmClient, a.cryptor)
		}

		if err != nil {
			session := sessions.Default(c)
			session.AddFlash(err.Error())
			err := session.Save()
			if err != nil {
				a.log.WithError(err).Error("problem saving session")
				c.Redirect(http.StatusSeeOther, fmt.Sprintf("/team/%v/%v/new", slug, chartType))
				return
			}
			c.Redirect(http.StatusSeeOther, fmt.Sprintf("/team/%v/%v/new", slug, chartType))
			return
		}

		c.Redirect(http.StatusSeeOther, "/user")
	})

	a.router.GET("/team/:team/:chart/edit", func(c *gin.Context) {
		slug := c.Param("team")
		chartType := getChartType(c.Param("chart"))

		team, err := a.repo.TeamGet(c, slug)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var form any
		switch chartType {
		case gensql.ChartTypeJupyterhub:
			form = &chart.JupyterConfigurableValues{}
		case gensql.ChartTypeAirflow:
			form = &chart.AirflowConfigurableValues{}
		}

		err = a.repo.TeamConfigurableValuesGet(c, chartType, team.ID, form)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		session := sessions.Default(c)
		flashes := session.Flashes()
		err = session.Save()
		if err != nil {
			a.log.WithError(err).Error("problem saving session")
			return
		}
		c.HTML(http.StatusOK, fmt.Sprintf("charts/%v", chartType), gin.H{
			"team":   slug,
			"values": form,
			"errors": flashes,
		})
	})

	a.router.POST("/team/:team/:chart/edit", func(c *gin.Context) {
		slug := c.Param("team")
		chartType := getChartType(c.Param("chart"))
		var err error

		switch chartType {
		case gensql.ChartTypeJupyterhub:
			var form chart.JupyterForm
			err = c.ShouldBindWith(&form, binding.Form)
			if err != nil {
				session := sessions.Default(c)
				session.AddFlash(err.Error())
				err := session.Save()
				if err != nil {
					a.log.WithError(err).Error("problem saving session")
					c.Redirect(http.StatusSeeOther, fmt.Sprintf("/team/%v/%v/edit", slug, chartType))
					return
				}
				c.Redirect(http.StatusSeeOther, fmt.Sprintf("/team/%v/%v/edit", slug, chartType))
				return
			}
			form.Slug = slug
			err = chart.UpdateJupyterhub(c, form, a.repo, a.helmClient, a.cryptor)
		case gensql.ChartTypeAirflow:
			var form chart.AirflowForm
			err = c.ShouldBindWith(&form, binding.Form)
			if err != nil {
				session := sessions.Default(c)
				session.AddFlash(err.Error())
				err := session.Save()
				if err != nil {
					a.log.WithError(err).Error("problem saving session")
					c.Redirect(http.StatusSeeOther, fmt.Sprintf("/team/%v/%v/edit", slug, chartType))
					return
				}
				c.Redirect(http.StatusSeeOther, fmt.Sprintf("/team/%v/%v/edit", slug, chartType))
				return
			}
			form.Slug = slug
			err = chart.UpdateAirflow(c, form, a.repo, a.helmClient, a.cryptor)
		}

		if err != nil {
			a.log.WithError(err).Errorf("problem editing chart %v for team %v", chartType, slug)
			session := sessions.Default(c)
			session.AddFlash(err.Error())
			err := session.Save()
			if err != nil {
				a.log.WithError(err).Error("problem saving session")
				c.Redirect(http.StatusSeeOther, fmt.Sprintf("/team/%v/%v/edit", slug, chartType))
				return
			}
			c.Redirect(http.StatusSeeOther, fmt.Sprintf("/team/%v/%v/edit", slug, chartType))
			return
		}
		c.Redirect(http.StatusSeeOther, "/user")
	})

	a.router.POST("/team/:team/:chart/delete", func(c *gin.Context) {
		slug := c.Param("team")
		chartType := getChartType(c.Param("chart"))
		var err error

		switch chartType {
		case gensql.ChartTypeJupyterhub:
			err = chart.DeleteJupyterhub(c, slug, a.repo, a.helmClient)
		case gensql.ChartTypeAirflow:
			err = chart.DeleteAirflow(c, slug, a.repo, a.helmClient, a.googleClient, a.k8sClient)
		}

		if err != nil {
			a.log.WithError(err).Errorf("problem deleting chart %v for team %v", chartType, slug)
			session := sessions.Default(c)
			session.AddFlash(err.Error())
			err := session.Save()
			if err != nil {
				a.log.WithError(err).Error("problem saving session")
				c.Redirect(http.StatusSeeOther, "/user")
				return
			}
			c.Redirect(http.StatusSeeOther, "/user")
			return
		}
		c.Redirect(http.StatusSeeOther, "/user")
	})
}
