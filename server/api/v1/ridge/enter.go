package ridge

import "github.com/springbear2020/self-hub/server/service"

type ApiGroup struct{ RidgeContactApi }

var ridgeContactService = service.ServiceGroupApp.RidgeServiceGroup.RidgeContactService
