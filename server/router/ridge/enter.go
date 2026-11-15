package ridge

import api "github.com/springbear2020/self-hub/server/api/v1"

type RouterGroup struct{ RidgeContactRouter }

var ridgeContactApi = api.ApiGroupApp.RidgeApiGroup.RidgeContactApi
