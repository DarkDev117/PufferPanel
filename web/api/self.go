/*
 Copyright 2020 Padduck, LLC
  Licensed under the Apache License, Version 2.0 (the "License");
  you may not use this file except in compliance with the License.
  You may obtain a copy of the License at
  	http://www.apache.org/licenses/LICENSE-2.0
  Unless required by applicable law or agreed to in writing, software
  distributed under the License is distributed on an "AS IS" BASIS,
  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
  See the License for the specific language governing permissions and
  limitations under the License.
*/

package api

import (
	"github.com/gin-gonic/gin"
	"github.com/pufferpanel/pufferpanel/v2"
	"github.com/pufferpanel/pufferpanel/v2/middleware"
	"github.com/pufferpanel/pufferpanel/v2/middleware/handlers"
	"github.com/pufferpanel/pufferpanel/v2/models"
	"github.com/pufferpanel/pufferpanel/v2/response"
	"github.com/pufferpanel/pufferpanel/v2/services"
	"net/http"
)

func registerSelf(g *gin.RouterGroup) {
	g.Handle("GET", "", handlers.HasOAuth2Token, getSelf)
	g.Handle("PUT", "", handlers.HasOAuth2Token, updateSelf)
}

func getSelf(c *gin.Context) {
	t, exist := c.Get("user")
	user, ok := t.(*models.User)

	if !exist || !ok {
		response.HandleError(c, pufferpanel.ErrUnknownError, http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, models.FromUser(user))
}

func updateSelf(c *gin.Context) {
	db := middleware.GetDatabase(c)
	us := &services.User{DB: db}

	t, exist := c.Get("user")
	user, ok := t.(*models.User)

	if !exist || !ok {
		response.HandleError(c, pufferpanel.ErrUnknownError, http.StatusInternalServerError)
		return
	}

	var viewModel models.UserView
	if err := c.BindJSON(&viewModel); response.HandleError(c, err, http.StatusBadRequest) {
		return
	}

	if err := viewModel.Valid(true); response.HandleError(c, err, http.StatusBadRequest) {
		return
	}

	if viewModel.Password == "" {
		response.HandleError(c, pufferpanel.ErrFieldRequired("password"), http.StatusBadRequest)
		return
	}

	if !us.IsValidCredentials(user, viewModel.Password) {
		response.HandleError(c, pufferpanel.ErrInvalidCredentials, http.StatusInternalServerError)
		return
	}

	viewModel.CopyToModel(user)

	if viewModel.NewPassword != "" {
		err := user.SetPassword(viewModel.NewPassword)
		if response.HandleError(c, err, http.StatusInternalServerError) {
			return
		}
	}

	if err := us.Update(user); response.HandleError(c, err, http.StatusInternalServerError) {
		return
	}

	c.Status(http.StatusNoContent)
}
