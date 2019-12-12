# Librería para notificaciones notificaciones_lib

La librería se encarga de enviar notificaciones que han sido configuradas previamente en [configuracion](https://pruebasconfiguracion.portaloas.udistrital.edu.co/#/pages/dashboard)


# Instalación e inicialización de la librería.

1. Importe la librería notificaciones_lib en main.go
```go
package main

import (
	_ "github.com/udistrital/configuracion_api/routers"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	"github.com/astaxie/beego/plugins/cors"
    ...
	notificacionlib "github.com/udistrital/notificaciones_lib"
    ...
)
```
2. Inicialize la librería en la funcion main() de main.go
```go

func main() {
	...
	notificacionlib.InitMiddleware()
    ..
	beego.Run()
}
```
