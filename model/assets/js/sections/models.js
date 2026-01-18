DOCS.models = {
  id: "models",
  title: "Defining Models",
  render(section) {
    section.append(
      UI.paragraph("Models are defined using Go structs and field builders."),
      UI.code(`
var Users = model.New(db, "users", struct {
  UserId *model.Field
}{
  UserId: model.CreateField().AsBigInt().IsPrimary(),
})
      `)
    )
  }
}
