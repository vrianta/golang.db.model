DOCS.queries = {
  id: "queries",
  title: "Query Builder",
  render(section) {
    section.append(
      UI.code(`
Users.Get().
  Where(Users.Fields.Email).Like("%gmail.com").
  Limit(10).
  Fetch()
      `)
    )
  }
}
