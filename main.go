package main

import (
	"bytes"
	"flag"
	"fmt"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"io/fs"
	"io/ioutil"
	"os"
	"sigs.k8s.io/yaml"
)

func Exit(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(0)
	}
}

func YamlToJson(filePath string) (jsons []byte) {
	file, err := ioutil.ReadFile(filePath)
	Exit(err)
	split := bytes.Split(file, []byte("---"))
	jsonData := []byte{}
	newLine := []byte("\n")
	for _, item := range split {
		if bytes.Equal(item, []byte("")) {
			continue
		}
		indent, _ := yaml.YAMLToJSON(item)
		jsonData = append(jsonData, indent...)
		jsonData = append(jsonData, newLine...)
	}

	Exit(err)
	return jsonData
}

func JsonToYaml(jsonData []byte, clusterName string) {
	yamlData, err := yaml.JSONToYAML(jsonData)
	if err != nil {
		fmt.Printf("err: %v\n", err)
		return
	}
	pwd, _ := os.Getwd()
	filename := fmt.Sprintf("/%v", clusterName)
	ioutil.WriteFile(pwd+filename, yamlData, fs.ModePerm)
	fmt.Println(string(yamlData))
}

// 按集群分割配置文件
func DisKubecfg(jsonData []byte) {
	// kubeconfig
	data := string(jsonData)
	// clusters
	clusters := gjson.Get(string(jsonData),"clusters")
	// contexts
	contexts := gjson.Get(string(jsonData),"contexts")
	// users
	users := gjson.Get(string(jsonData),"users")
	i := 0
	clusters.ForEach(func(_, c gjson.Result) bool {
		clustername := gjson.Get(c.String(),"name")
		data,_ = sjson.Set(data,"clusters",c.Value() )
		// contexts handle
		contexts_list := `{}`
		users_list := `{}`
		j := 0
		contexts.ForEach(func(_, ctx gjson.Result) bool {
			ctxc := gjson.Get(ctx.String(), "context.cluster")
			ctxn := gjson.Get(ctx.String(), "name")
			if ctxc.String() == clustername.String() {
				if i==0 || contexts_list != `{}` {
					contexts_list,_ = sjson.Set(contexts_list,fmt.Sprintf("contexts.%v", i),ctx.Value() )
					i++
					// users handle
					ctxun := gjson.Get(ctx.String(), "context.user")
					users.ForEach(func(_, u gjson.Result) bool {
						username := gjson.Get(u.String(),"name")
						if ctxun.String() == username.String() && (j==0 || users_list != `{})`) {
							users_list,_ = sjson.Set(users_list,fmt.Sprintf("users.%v", j),u.Value() )
							j++
						}
						return true
					})
					u_list := gjson.Get(string(users_list),"users")
					data,_ = sjson.Set(data,"users",u_list.Value())
					// current-context handle
					data,_ = sjson.Set(data,"current-context",ctxn.Value() )
				}
			}
			ctx_list := gjson.Get(string(contexts_list),"contexts")
			data,_ = sjson.Set(data,"contexts",ctx_list.Value())

			return true
		})
		i=0
		j=0
		JsonToYaml([]byte(data), clustername.String())
		return true
	})

}

var (
	kubeconfig string //kubeconfig
)

func init() {
	home := os.Getenv("HOME")
	flag.StringVar(&kubeconfig, "f", fmt.Sprintf("%v/.kube/config",home), "./dis -f config")
}

func main() {
	flag.Parse()
	jsondata := YamlToJson(kubeconfig)
	DisKubecfg(jsondata)
}