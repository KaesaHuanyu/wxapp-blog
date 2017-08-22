package handler

import (
	"github.com/labstack/echo"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"time"
)

func (h *handler) AdminInterface(c echo.Context) error {
	body, err := ioutil.ReadFile(`view/adminInterface.html`)
	if err != nil {
		logrus.Errorf("[%s] AdminInterface ReadFile error: [%s]", time.Now().String(), err.Error())
		return c.NoContent(http.StatusInternalServerError)
	}
	return c.HTML(http.StatusOK, string(body))
}
