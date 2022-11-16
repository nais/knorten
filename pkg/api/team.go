package api

import (
	"fmt"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/nais/knorten/pkg/team"
	"net/http"
)

func (a *API) setupTeamRoutes() {
	a.router.GET("/team/new", func(c *gin.Context) {
		var form team.Form
		session := sessions.Default(c)
		flashes := session.Flashes()
		err := session.Save()
		if err != nil {
			a.log.WithError(err).Error("problem saving session")
			return
		}

		c.HTML(http.StatusOK, "team/new", gin.H{
			"form":   form,
			"errors": flashes,
		})
	})

	a.router.POST("/team/new", func(c *gin.Context) {
		err := team.Create(c, a.repo, a.googleClient, a.k8sClient)
		if err != nil {
			session := sessions.Default(c)
			session.AddFlash(err.Error())
			err := session.Save()
			if err != nil {
				a.log.WithError(err).Error("problem saving session")
				return
			}
			c.Redirect(http.StatusSeeOther, "/team/new")
			return
		}
		c.Redirect(http.StatusSeeOther, "/user")
	})

	a.router.GET("/team/:team/edit", func(c *gin.Context) {
		teamName := c.Param("team")
		get, err := a.repo.TeamGet(c, teamName)
		if err != nil {
			a.log.WithError(err).Errorf("problem getting team %v", teamName)
			c.Redirect(http.StatusSeeOther, "/user")
			return
		}

		session := sessions.Default(c)
		flashes := session.Flashes()
		err = session.Save()
		if err != nil {
			a.log.WithError(err).Error("problem saving session")
			return
		}
		c.HTML(http.StatusOK, "team/edit", gin.H{
			"users":  get.Users,
			"team":   teamName,
			"errors": flashes,
		})
	})

	a.router.POST("/team/:team/edit", func(c *gin.Context) {
		teamName := c.Param("team")
		err := team.Update(c, a.repo, a.googleClient, a.helmClient)
		if err != nil {
			session := sessions.Default(c)
			session.AddFlash(err.Error())
			err := session.Save()
			if err != nil {
				a.log.WithError(err).Error("problem saving session")
				return
			}
			c.Redirect(http.StatusSeeOther, fmt.Sprintf("/team/%v/edit", teamName))
			return
		}
		c.Redirect(http.StatusSeeOther, "/user")
	})
}