package controllers

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/astaxie/beego"

	"github.com/dockercn/wharf/models"
	"github.com/dockercn/wharf/modules"
)

type ManifestsAPIV2Controller struct {
	beego.Controller
}

func (this *ManifestsAPIV2Controller) URLMapping() {
}

func (this *ManifestsAPIV2Controller) Prepare() {
	this.EnableXSRF = false

	this.Ctx.Output.Context.ResponseWriter.Header().Set("Content-Type", "application/json;charset=UTF-8")
}

func (this *ManifestsAPIV2Controller) PutManifests() {
	manifest, _ := ioutil.ReadAll(this.Ctx.Request.Body)

	namespace := this.Ctx.Input.Param(":namespace")
	repository := this.Ctx.Input.Param(":repo_name")

	repo := new(models.Repository)

	if err := repo.Put(namespace, repository, "", this.Ctx.Input.Header("User-Agent"), models.APIVERSION_V2); err != nil {
		result := map[string][]modules.ErrorDescriptor{"errors": []modules.ErrorDescriptor{modules.ErrorDescriptors[modules.APIErrorCodeManifestInvalid]}}
		this.Data["json"] = &result

		this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
		this.ServeJson()
		return
	}

	if err := manifestsConvertV1(manifest); err != nil {
		beego.Error("[REGISTRY API V2] Decode Manifest Error: ", err.Error())
	}

	this.Ctx.Output.Context.Output.SetStatus(http.StatusOK)
	this.Ctx.Output.Context.Output.Body([]byte(""))
	return
}

func (this *ManifestsAPIV2Controller) GetTags() {
	namespace := this.Ctx.Input.Param(":namespace")
	repository := this.Ctx.Input.Param(":repo_name")

	repo := new(models.Repository)

	if has, _, err := repo.Has(namespace, repository); err != nil || has == false {
		result := map[string][]modules.ErrorDescriptor{"errors": []modules.ErrorDescriptor{modules.ErrorDescriptors[modules.APIErrorCodeNameInvalid]}}
		this.Data["json"] = &result

		this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
		this.ServeJson()
		return
	}

	data := map[string]interface{}{}
	tags := []string{}

	data["name"] = fmt.Sprintf("%s/%s", namespace, repository)

	for _, value := range repo.Tags {
		t := new(models.Tag)
		if err := t.GetByUUID(value); err != nil {
			result := map[string][]modules.ErrorDescriptor{"errors": []modules.ErrorDescriptor{modules.ErrorDescriptors[modules.APIErrorCodeTagInvalid]}}
			this.Data["json"] = &result

			this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
			this.ServeJson()
			return
		}

		tags = append(tags, t.Name)
	}

	data["tags"] = tags
	this.Data["json"] = data

	this.Ctx.Output.Context.Output.SetStatus(http.StatusOK)
	this.ServeJson()
	return
}

func (this *ManifestsAPIV2Controller) GetManifests() {
	namespace := this.Ctx.Input.Param(":namespace")
	repository := this.Ctx.Input.Param(":repo_name")
	tag := this.Ctx.Input.Param(":tag")

	t := new(models.Tag)
	if err := t.GetByUUID(fmt.Sprintf("%s:%s:%s", namespace, repository, tag)); err != nil {
		result := map[string][]modules.ErrorDescriptor{"errors": []modules.ErrorDescriptor{modules.ErrorDescriptors[modules.APIErrorCodeTagInvalid]}}
		this.Data["json"] = &result

		this.Ctx.Output.Context.Output.SetStatus(http.StatusBadRequest)
		this.ServeJson()
		return
	}

	beego.Trace("[Docker Registry API V2] Manifests:", t.Manifest)

	this.Ctx.Output.Context.Output.SetStatus(http.StatusOK)
	this.Ctx.Output.Context.Output.Body([]byte(t.Manifest))
	return
}
