const nav = document.getElementById("nav")
const content = document.getElementById("content")

DOC_ORDER.forEach(key => {
  const doc = DOCS[key]
  if (!doc) return

  // nav
  const link = document.createElement("a")
  link.className = "nav-link"
  link.href = `#${doc.id}`
  link.textContent = doc.title
  nav.appendChild(link)

  // section
  const section = UI.section(doc.id)
  section.append(UI.title(doc.title))
  doc.render(section)

  content.appendChild(section)
})

// scroll spy
const links = document.querySelectorAll(".nav-link")
window.addEventListener("scroll", () => {
  const y = window.scrollY + 120
  links.forEach(l => {
    const s = document.querySelector(l.getAttribute("href"))
    l.classList.toggle(
      "active",
      s.offsetTop <= y && s.offsetTop + s.offsetHeight > y
    )
  })
})
