package notification

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/context"
	"github.com/astaxie/beego/httplib"
)

type Mensaje struct {
	Message string
}

func failOnError(err error, msg string) {
	if err != nil {
		beego.Info("%s: %s", msg, err)
		beego.Info(fmt.Sprintf("%s: %s", msg, err))
	}
}

func FunctionBeforeStatic(ctx *context.Context) {
	beego.Info("beego.BeforeStatic: Before finding the static file")
}
func FunctionBeforeRouter(ctx *context.Context) {
	beego.Info("beego.BeforeRouter: Executing Before finding router")
}
func FunctionBeforeExec(ctx *context.Context) {
	var u map[string]interface{}
	FillStruct(ctx.Input.Data()["json"], &u)
	beego.Info(u)
	beego.Info("beego.BeforeExec: After finding router and before executing the matched Controller")
}

func FunctionAfterExec(ctx *context.Context) {
	var mt Mensaje
	var jsonMensaje string
	var jsonBytes []byte
	var res interface{}
	var v []map[string]interface{}
	var u map[string]interface{}
	var x map[string]interface{}
	json.Unmarshal(ctx.Input.RequestBody, &x)
	FillStruct(ctx.Input.Data()["json"], &u)
	if tip, e := u["Type"].(string); e {
		r := httplib.Get(beego.AppConfig.String("configuracionService") + "notificacion_configuracion")
		beego.Info(r)
		r.Param("query", "EndPoint:"+ctx.Request.URL.String()+",MetodoHttp.Nombre:"+ctx.Request.Method+",Tipo.Nombre:"+tip+",Aplicacion.Nombre:"+
			beego.AppConfig.String("appname"))
		if err := r.ToJSON(&v); err == nil && v != nil {
			if NotConf, err := profilesExtract(v[0]); err == nil {
				jsonMensaje = NotConf["CuerpoNotificacion"].(string)
				beego.Info("jsonMensaje:" + jsonMensaje)
				json.Unmarshal(jsonBytes, &mt.Message)
				app := NotConf["Aplicacion"].(map[string]interface{})
				mt.Message = formatNotificationMessage(mt.Message, u)
				NotConf["CuerpoNotificacion"] = mt
				data := make(map[string]interface{})

				if x["NotifyUser"] == nil {
					data = map[string]interface{}{
						"ConfiguracionNotificacion": NotConf["Id"],
						"DestinationProfiles":       NotConf["Perfiles"],
						"Application":               NotConf["App"],
						"NotificationBody":          NotConf["CuerpoNotificacion"],
						"UserDestination":           "",
						"Alias":                     app["Alias"],
						"EstiloIcono":               app["EstiloIcono"],
					}
				} else {
					data = map[string]interface{}{
						"ConfiguracionNotificacion": NotConf["Id"],
						"DestinationProfiles":       nil,
						"Application":               NotConf["App"],
						"NotificationBody":          NotConf["CuerpoNotificacion"],
						"UserDestination":           x["NotifyUser"],
						"Alias":                     app["Alias"],
						"EstiloIcono":               app["EstiloIcono"]}
				}
				beego.Info("Test:" + beego.AppConfig.String("notificacionService") + "notify")
				sendJson(beego.AppConfig.String("notificacionService")+"notify", "POST", &res, data)
			}
		} else {
			beego.Info(err)
		}

	}

}
func profilesExtract(configData map[string]interface{}) (conf map[string]interface{}, err error) {
	var auxStr string
	var profileConf []map[string]interface{}
	var profiles []string
	conf = configData
	if err = FillStructDeep(configData, "NotificacionConfiguracionPerfil", &profileConf); err == nil {
		for _, data := range profileConf {
			if err = FillStructDeep(data, "Perfil.Nombre", &auxStr); err == nil {
				profiles = append(profiles, auxStr)
			} else {
				return
			}
		}
	} else {
		return
	}
	conf["Perfiles"] = profiles
	var aux interface{}
	FillStructDeep(conf, "Aplicacion.Nombre", &aux)
	conf["App"] = aux
	return
}

func formatNotificationMessage(message string, data map[string]interface{}) (res string) {
	res = message
	var deepData interface{}
	r, _ := regexp.Compile("<field>([a-zA-Z.]+)</field>")
	fields := r.FindAllStringSubmatch(message, -1)
	for _, field := range fields {
		FillStructDeep(data, field[1], &deepData)
		textReplace := fmt.Sprintf("%v", deepData)
		res = strings.Replace(res, "<field>"+field[1]+"</field>", textReplace, -1)
	}

	return
}

func FillStruct(m interface{}, s interface{}) (err error) {
	j, _ := json.Marshal(m)
	err = json.Unmarshal(j, s)
	return
}

func FillStructDeep(m map[string]interface{}, fields string, s interface{}) (err error) {
	f := strings.Split(fields, ".")
	if len(f) == 0 {
		err = errors.New("invalid fields.")
		return
	}

	var aux map[string]interface{}
	var load interface{}
	for i, value := range f {

		if i == 0 {
			//fmt.Println(m[value])
			FillStruct(m[value], &load)
		} else {
			FillStruct(load, &aux)
			FillStruct(aux[value], &load)
			//fmt.Println(aux[value])
		}
	}
	j, _ := json.Marshal(load)
	err = json.Unmarshal(j, s)
	return
}

func FunctionFinishRouter(ctx *context.Context) {
	beego.Info("beego.FinishRouter: After finishing router")
}

func InitMiddleware() {
	beego.Info("init...")
	beego.InsertFilter("*", beego.AfterExec, FunctionAfterExec, false)
}

func sendJson(urlp string, trequest string, target interface{}, datajson interface{}) error {
	b := new(bytes.Buffer)
	if datajson != nil {
		json.NewEncoder(b).Encode(datajson)
	}
	//proxyUrl, err := url.Parse("http://10.20.4.15:3128")
	//http.DefaultTransport = &http.Transport{Proxy: http.ProxyURL(proxyUrl)}
	client := &http.Client{}
	req, err := http.NewRequest(trequest, urlp, b)
	r, err := client.Do(req)
	//r, err := http.Post(url, "application/json; charset=utf-8", b)
	if err != nil {
		beego.Error("error", err)
		return err
	}
	defer r.Body.Close()

	return json.NewDecoder(r.Body).Decode(target)
}

func getJson(urlp string, target interface{}) error {
	//proxyUrl, err := url.Parse("http://10.20.4.15:3128")
	//http.DefaultTransport = &http.Transport{Proxy: http.ProxyURL(proxyUrl)}
	r, err := http.Get(urlp)
	if err != nil {
		return err
	}
	defer r.Body.Close()

	return json.NewDecoder(r.Body).Decode(target)
}
