window.UI = {
  section(id) {
    const el = document.createElement("section")
    el.className = "doc-section"
    el.id = id
    return el
  },

  title(text) {
    const h = document.createElement("h2")
    h.className = "doc-title"
    h.textContent = text
    return h
  },

  paragraph(text) {
    const p = document.createElement("p")
    p.className = "doc-text"
    p.innerHTML = text
    return p
  },

  code(code) {
    const pre = document.createElement("pre")
    const c = document.createElement("code")
    c.textContent = code.trim()
    pre.appendChild(c)
    return pre
  },

  list(items) {
    const ul = document.createElement("ul")
    ul.className = "doc-text"
    items.forEach(i => {
      const li = document.createElement("li")
      li.textContent = i
      ul.appendChild(li)
    })
    return ul
  }
}
