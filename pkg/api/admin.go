package api

import (
	"encoding/gob"
	"fmt"
	"github.com/nais/knorten/pkg/database/gensql"
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

type teamInfo struct {
	gensql.Team
	Apps []string
}

func (a *API) setupAdminRoutes() {
	a.router.GET("/admin", func(c *gin.Context) {
		session := sessions.Default(c)
		flashes := session.Flashes()
		err := session.Save()
		if err != nil {
			a.log.WithError(err).Error("problem saving session")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err})
			return
		}

		teams, err := a.repo.TeamsGet(c)
		if err != nil {
			session := sessions.Default(c)
			session.AddFlash(err.Error())
			err = session.Save()
			if err != nil {
				a.log.WithError(err).Error("problem saving session")
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err})
				return
			}

			c.Redirect(http.StatusSeeOther, "/admin")
			return
		}

		teamApps := map[string]teamInfo{}
		for _, team := range teams {
			apps, err := a.repo.AppsForTeamGet(c, team.ID)
			if err != nil {
				a.log.WithError(err).Error("problem retrieving apps for teams")
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err})
				return
			}
			teamApps[team.ID] = teamInfo{
				Team: team,
				Apps: apps,
			}
		}

		c.HTML(http.StatusOK, "admin/index", gin.H{
			"errors": flashes,
			"teams":  teamApps,
		})
	})

	a.router.GET("/admin/:chart", func(c *gin.Context) {
		chartType := getChartType(c.Param("chart"))

		values, err := a.repo.GlobalValuesGet(c, chartType)
		if err != nil {
			session := sessions.Default(c)
			session.AddFlash(err.Error())
			err = session.Save()
			if err != nil {
				a.log.WithError(err).Error("problem saving session")
				c.Redirect(http.StatusSeeOther, "/admin")
				return
			}

			c.Redirect(http.StatusSeeOther, "/admin")
			return
		}

		session := sessions.Default(c)
		flashes := session.Flashes()
		err = session.Save()
		if err != nil {
			a.log.WithError(err).Error("problem saving session")
			c.Redirect(http.StatusSeeOther, "/admin")
			return
		}

		c.HTML(http.StatusOK, "admin/chart", gin.H{
			"values": values,
			"errors": flashes,
			"chart":  string(chartType),
		})
	})

	a.router.POST("/admin/:chart", func(c *gin.Context) {
		session := sessions.Default(c)
		chartType := getChartType(c.Param("chart"))

		err := c.Request.ParseForm()
		if err != nil {
			session.AddFlash(err.Error())
			err = session.Save()
			if err != nil {
				a.log.WithError(err).Error("problem saving session")
				c.Redirect(http.StatusSeeOther, "admin")
				return
			}
			c.Redirect(http.StatusSeeOther, "admin")
			return
		}

		changedValues, err := a.adminClient.FindGlobalValueChanges(c, c.Request.PostForm, chartType)
		if err != nil {
			session := sessions.Default(c)
			session.AddFlash(err.Error())
			err = session.Save()
			if err != nil {
				a.log.WithError(err).Error("problem saving session")
				c.Redirect(http.StatusSeeOther, fmt.Sprintf("/admin/%v", chartType))
				return
			}

			c.Redirect(http.StatusSeeOther, fmt.Sprintf("/admin/%v", chartType))
			return
		}

		if len(changedValues) == 0 {
			session.AddFlash("Ingen endringer lagret")
			err = session.Save()
			if err != nil {
				a.log.WithError(err).Error("problem saving session")
				c.Redirect(http.StatusSeeOther, fmt.Sprintf("/admin/%v", chartType))
				return
			}
			c.Redirect(http.StatusSeeOther, "/admin")
			return
		}

		gob.Register(changedValues)
		session.AddFlash(changedValues)
		err = session.Save()
		if err != nil {
			a.log.WithError(err).Error("problem saving session")
			c.Redirect(http.StatusSeeOther, fmt.Sprintf("/admin/%v", chartType))
			return
		}
		c.Redirect(http.StatusSeeOther, fmt.Sprintf("/admin/%v/confirm", chartType))
	})

	a.router.GET("/admin/:chart/confirm", func(c *gin.Context) {
		chartType := getChartType(c.Param("chart"))
		session := sessions.Default(c)
		changedValues := session.Flashes()
		err := session.Save()
		if err != nil {
			a.log.WithError(err).Error("problem saving session")
			c.Redirect(http.StatusSeeOther, fmt.Sprintf("/admin/%v", chartType))
			return
		}

		c.HTML(http.StatusOK, "admin/confirm", gin.H{
			"changedValues": changedValues,
			"chart":         string(chartType),
		})
	})

	a.router.POST("/admin/:chart/confirm", func(c *gin.Context) {
		session := sessions.Default(c)
		chartType := getChartType(c.Param("chart"))

		err := c.Request.ParseForm()
		if err != nil {
			a.log.WithError(err)
			session.AddFlash(err.Error())
			err = session.Save()
			if err != nil {
				a.log.WithError(err).Error("problem saving session")
				c.Redirect(http.StatusSeeOther, fmt.Sprintf("/admin/%v/confirm", chartType))
				return
			}
			c.Redirect(http.StatusSeeOther, fmt.Sprintf("/admin/%v/confirm", chartType))
			return
		}

		if err := a.adminClient.UpdateGlobalValues(c, c.Request.PostForm, chartType); err != nil {
			c.Redirect(http.StatusSeeOther, "/admin")
			return
		}

		if err != nil {
			a.log.WithError(err)
			session.AddFlash(err.Error())
			err = session.Save()
			if err != nil {
				a.log.WithError(err).Error("problem saving session")
				c.Redirect(http.StatusSeeOther, fmt.Sprintf("/admin/%v/confirm", chartType))
				return
			}
			c.Redirect(http.StatusSeeOther, fmt.Sprintf("/admin/%v/confirm", chartType))
			return
		}

		c.Redirect(http.StatusSeeOther, "/admin")
	})

	a.router.POST("/admin/:chart/sync", func(c *gin.Context) {
		chartType := getChartType(c.Param("chart"))
		team := c.PostForm("team")
		switch chartType {
		case gensql.ChartTypeJupyterhub:
			a.chartClient.Jupyterhub.Sync(c, team)
		case gensql.ChartTypeAirflow:
			a.chartClient.Airflow.Sync(c, team)
		}

		c.Redirect(http.StatusSeeOther, "/admin")
	})
}
