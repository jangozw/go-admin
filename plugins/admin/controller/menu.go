package controller

import (
	"encoding/json"
	"github.com/chenhg5/go-admin/context"
	"github.com/chenhg5/go-admin/modules/auth"
	"github.com/chenhg5/go-admin/modules/menu"
	"github.com/chenhg5/go-admin/plugins/admin/models"
	"github.com/chenhg5/go-admin/plugins/admin/modules/constant"
	"github.com/chenhg5/go-admin/plugins/admin/modules/guard"
	"github.com/chenhg5/go-admin/plugins/admin/modules/response"
	"github.com/chenhg5/go-admin/plugins/admin/modules/table"
	"github.com/chenhg5/go-admin/template"
	"github.com/chenhg5/go-admin/template/types"
	template2 "html/template"
	"net/http"
)

func ShowMenu(ctx *context.Context) {
	getMenuInfoPanel(ctx, "")
	return
}

func ShowEditMenu(ctx *context.Context) {

	formData, title, description := table.List["menu"].GetDataFromDatabaseWithId(ctx.Query("id"))

	user := auth.Auth(ctx)

	js := `<script>
$('.icon').iconpicker({placement: 'bottomLeft'});
</script>`

	tmpl, tmplName := aTemplate().GetTemplate(isPjax(ctx))
	buf := template.Execute(tmpl, tmplName, user, types.Panel{
		Content: aForm().
			SetContent(formData).
			SetPrefix(config.PrefixFixSlash()).
			SetUrl(config.Url("/menu/edit")).
			SetToken(auth.TokenHelper.AddToken()).
			SetInfoUrl(config.Url("/menu")).
			GetContent() + template2.HTML(js),
		Description: description,
		Title:       title,
	}, config, menu.GetGlobalMenu(user).SetActiveClass(config.UrlRemovePrefix(ctx.Path())))

	ctx.Html(http.StatusOK, buf.String())
}

func DeleteMenu(ctx *context.Context) {

	models.MenuWithId(guard.GetMenuDeleteParam(ctx).Id).Delete()
	menu.SetGlobalMenu(auth.Auth(ctx).WithRoles().WithMenus())
	table.RefreshTableList()
	response.Ok(ctx)
}

func EditMenu(ctx *context.Context) {

	param := guard.GetMenuEditParam(ctx)

	if param.HasAlert() {
		getMenuInfoPanel(ctx, param.Alert)
		ctx.AddHeader("Content-Type", "text/html; charset=utf-8")
		ctx.AddHeader(constant.PjaxUrlHeader, config.Url("/menu"))
		return
	}

	menuModel := models.MenuWithId(param.Id)

	for _, roleId := range param.Roles {
		menuModel.AddRole(roleId)
	}

	menuModel.Update(param.Title, param.ParentId, param.Icon, param.Uri)

	menu.SetGlobalMenu(auth.Auth(ctx))

	getMenuInfoPanel(ctx, "")
	ctx.AddHeader("Content-Type", "text/html; charset=utf-8")
	ctx.AddHeader(constant.PjaxUrlHeader, config.Url("/menu"))
}

func NewMenu(ctx *context.Context) {

	param := guard.GetMenuNewParam(ctx)

	if param.HasAlert() {
		getMenuInfoPanel(ctx, param.Alert)
		ctx.AddHeader("Content-Type", "text/html; charset=utf-8")
		ctx.AddHeader(constant.PjaxUrlHeader, config.Url("/menu"))
		return
	}

	user := auth.Auth(ctx)

	menuModel := models.Menu().New(param.Title, param.ParentId, param.Icon, param.Uri, (menu.GetGlobalMenu(user)).MaxOrder+1)

	for _, roleId := range param.Roles {
		menuModel.AddRole(roleId)
	}

	menu.GetGlobalMenu(user.WithRoles().WithMenus()).AddMaxOrder()
	table.RefreshTableList()

	getMenuInfoPanel(ctx, "")
	ctx.AddHeader("Content-Type", "text/html; charset=utf-8")
	ctx.AddHeader(constant.PjaxUrlHeader, config.Url("/menu"))
}

func MenuOrder(ctx *context.Context) {

	var data []map[string]interface{}
	_ = json.Unmarshal([]byte(ctx.FormValue("_order")), &data)

	models.Menu().ResetOrder(data)
	menu.SetGlobalMenu(auth.Auth(ctx))

	response.Ok(ctx)
	return
}

func getMenuInfoPanel(ctx *context.Context, alert template2.HTML) {
	user := auth.Auth(ctx)

	menu.GlobalMenu.SetActiveClass(config.UrlRemovePrefix(ctx.Path()))

	editUrl := config.Url("/menu/edit/show")
	deleteUrl := config.Url("/menu/delete")
	orderUrl := config.Url("/menu/order")

	tree := aTree().
		SetTree((menu.GetGlobalMenu(user)).GlobalMenuList).
		SetEditUrl(editUrl).
		SetDeleteUrl(deleteUrl).
		SetOrderUrl(orderUrl).
		GetContent()

	header := aTree().GetTreeHeader()
	box := aBox().SetHeader(header).SetBody(tree).GetContent()
	col1 := aCol().SetSize(map[string]string{"md": "6"}).SetContent(box).GetContent()

	newForm := aForm().
		SetPrefix(config.PrefixFixSlash()).
		SetUrl(config.Url("/menu/new")).
		SetInfoUrl(config.Url("/menu")).
		SetTitle("New").
		SetContent(table.GetNewFormList(table.List["menu"].GetForm().FormList)).
		GetContent()

	col2 := aCol().SetSize(map[string]string{"md": "6"}).SetContent(newForm).GetContent()

	row := aRow().SetContent(col1 + col2).GetContent()

	menu.GlobalMenu.SetActiveClass(config.UrlRemovePrefix(ctx.Path()))

	tmpl, tmplName := aTemplate().GetTemplate(isPjax(ctx))
	buf := template.Execute(tmpl, tmplName, user, types.Panel{
		Content:     alert + row,
		Description: "Menus Manage",
		Title:       "Menus Manage",
	}, config, menu.GetGlobalMenu(user).SetActiveClass(config.UrlRemovePrefix(ctx.Path())))

	ctx.Html(http.StatusOK, buf.String())
}
