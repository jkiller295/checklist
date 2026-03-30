package handlers

import (
	"checklist/internal/db"
	"checklist/internal/i18n"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type Server struct {
	DB       *db.DB
	Password string
	Tmpls    *template.Template
}

// ─── helpers ────────────────────────────────────────────────────────────────

func (s *Server) lang(r *http.Request) string {
	// 1. cookie
	if c, err := r.Cookie("lang"); err == nil && i18n.ValidLang(c.Value) {
		return i18n.Normalize(c.Value)
	}
	// 2. Accept-Language header
	al := r.Header.Get("Accept-Language")
	for _, part := range strings.Split(al, ",") {
		tag := strings.TrimSpace(strings.Split(part, ";")[0])
		if i18n.ValidLang(tag) {
			return i18n.Normalize(tag)
		}
	}
	return "en"
}

func (s *Server) t(r *http.Request, key string) string {
	return i18n.T(s.lang(r), key)
}

func (s *Server) authed(r *http.Request) bool {
	c, err := r.Cookie("session")
	return err == nil && c.Value == "ok"
}

func (s *Server) render(w http.ResponseWriter, name string, data any) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := s.Tmpls.ExecuteTemplate(w, name, data); err != nil {
		log.Printf("template %s: %v", name, err)
		http.Error(w, "render error", 500)
	}
}

// ─── auth middleware ─────────────────────────────────────────────────────────

func (s *Server) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/login" || strings.HasPrefix(r.URL.Path, "/static/") {
			next.ServeHTTP(w, r)
			return
		}
		if !s.authed(r) {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// ─── routes ──────────────────────────────────────────────────────────────────

func (s *Server) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /login", s.handleLoginPage)
	mux.HandleFunc("POST /login", s.handleLogin)
	mux.HandleFunc("POST /logout", s.handleLogout)
	mux.HandleFunc("POST /lang", s.handleLang)

	mux.HandleFunc("GET /", s.handleHome)
	mux.HandleFunc("POST /lists", s.handleCreateList)
	mux.HandleFunc("DELETE /lists/{id}", s.handleDeleteList)
	mux.HandleFunc("PUT /lists/{id}", s.handleRenameList)

	mux.HandleFunc("GET /lists/{id}", s.handleListPage)
	mux.HandleFunc("GET /lists/{id}/stats", s.handleListStats)
	mux.HandleFunc("POST /lists/{id}/items", s.handleCreateItem)
	mux.HandleFunc("PATCH /items/{id}/toggle", s.handleToggleItem)
	mux.HandleFunc("PUT /items/{id}", s.handleUpdateItem)
	mux.HandleFunc("DELETE /items/{id}", s.handleDeleteItem)
	mux.HandleFunc("DELETE /lists/{id}/checked", s.handleClearChecked)
	mux.HandleFunc("POST /lists/{id}/check-all", s.handleCheckAll)
	mux.HandleFunc("POST /lists/{id}/uncheck-all", s.handleUncheckAll)
}

// ─── auth ────────────────────────────────────────────────────────────────────

func (s *Server) handleLoginPage(w http.ResponseWriter, r *http.Request) {
	s.render(w, "login.html", map[string]any{
		"T":    func(k string) string { return s.t(r, k) },
		"Lang": s.lang(r),
	})
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	pw := r.FormValue("password")
	if pw != s.Password {
		w.WriteHeader(http.StatusUnauthorized)
		s.render(w, "login.html", map[string]any{
			"T":     func(k string) string { return s.t(r, k) },
			"Lang":  s.lang(r),
			"Error": s.t(r, "wrong_password"),
		})
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name: "session", Value: "ok",
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int((30 * 24 * time.Hour).Seconds()),
	})
	http.Redirect(w, r, "/", http.StatusFound)
}

func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{Name: "session", Value: "", Path: "/", MaxAge: -1})
	http.Redirect(w, r, "/login", http.StatusFound)
}

func (s *Server) handleLang(w http.ResponseWriter, r *http.Request) {
	lang := r.FormValue("lang")
	if !i18n.ValidLang(lang) {
		lang = "en"
	}
	http.SetCookie(w, &http.Cookie{
		Name: "lang", Value: i18n.Normalize(lang),
		Path: "/", MaxAge: int((365 * 24 * time.Hour).Seconds()),
	})
	ref := r.Header.Get("Referer")
	if ref == "" {
		ref = "/"
	}
	// Support both HTMX and plain form requests
	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("HX-Redirect", ref)
		w.WriteHeader(http.StatusNoContent)
	} else {
		http.Redirect(w, r, ref, http.StatusFound)
	}
}

// ─── lists ───────────────────────────────────────────────────────────────────

func (s *Server) handleHome(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	lists, err := s.DB.GetLists()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	s.render(w, "home-page", map[string]any{
		"T":     func(k string) string { return s.t(r, k) },
		"Lang":  s.lang(r),
		"Lists": lists,
	})
}

func (s *Server) handleCreateList(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimSpace(r.FormValue("name"))
	if name == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	list, err := s.DB.CreateList(name)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	s.render(w, "list-card.html", map[string]any{
		"T":    func(k string) string { return s.t(r, k) },
		"Lang": s.lang(r),
		"List": list,
	})
}

func (s *Server) handleDeleteList(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err := s.DB.DeleteList(id); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleRenameList(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	name := strings.TrimSpace(r.FormValue("name"))
	if name == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if err := s.DB.RenameList(id, name); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	list, err := s.DB.GetList(id)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	s.render(w, "list-card.html", map[string]any{
		"T":    func(k string) string { return s.t(r, k) },
		"Lang": s.lang(r),
		"List": list,
	})
}

// ─── items ────────────────────────────────────────────────────────────────────

func (s *Server) handleListPage(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	list, err := s.DB.GetList(id)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	items, err := s.DB.GetItems(id)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	s.render(w, "list-page", map[string]any{
		"T":     func(k string) string { return s.t(r, k) },
		"Lang":  s.lang(r),
		"List":  list,
		"Items": items,
	})
}

func (s *Server) handleCreateItem(w http.ResponseWriter, r *http.Request) {
	listID, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	text := strings.TrimSpace(r.FormValue("text"))
	if text == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	item, err := s.DB.CreateItem(listID, text)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	s.render(w, "item-row.html", map[string]any{
		"T":    func(k string) string { return s.t(r, k) },
		"Lang": s.lang(r),
		"Item": item,
	})
}

func (s *Server) handleToggleItem(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	item, err := s.DB.ToggleItem(id)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	s.render(w, "item-row.html", map[string]any{
		"T":    func(k string) string { return s.t(r, k) },
		"Lang": s.lang(r),
		"Item": item,
	})
}

func (s *Server) handleUpdateItem(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	text := strings.TrimSpace(r.FormValue("text"))
	if text == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	item, err := s.DB.UpdateItem(id, text)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	s.render(w, "item-row.html", map[string]any{
		"T":    func(k string) string { return s.t(r, k) },
		"Lang": s.lang(r),
		"Item": item,
	})
}

func (s *Server) handleDeleteItem(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err := s.DB.DeleteItem(id); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleClearChecked(w http.ResponseWriter, r *http.Request) {
	listID, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err := s.DB.ClearChecked(listID); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	// Return updated item list
	items, err := s.DB.GetItems(listID)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	list, err := s.DB.GetList(listID)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	s.render(w, "item-list.html", map[string]any{
		"T":     func(k string) string { return s.t(r, k) },
		"Lang":  s.lang(r),
		"List":  list,
		"Items": items,
	})
}

func (s *Server) handleListStats(w http.ResponseWriter, r *http.Request) {
	listID, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	list, err := s.DB.GetList(listID)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	s.render(w, "list-stats.html", map[string]any{
		"T":    func(k string) string { return s.t(r, k) },
		"Lang": s.lang(r),
		"List": list,
	})
}

func (s *Server) handleCheckAll(w http.ResponseWriter, r *http.Request) {
	listID, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err := s.DB.CheckAll(listID); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	items, err := s.DB.GetItems(listID)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	list, err := s.DB.GetList(listID)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	// Return refreshed item list; stats refresh via htmx afterSwap
	s.render(w, "item-list.html", map[string]any{
		"T":     func(k string) string { return s.t(r, k) },
		"Lang":  s.lang(r),
		"List":  list,
		"Items": items,
	})
}

func (s *Server) handleUncheckAll(w http.ResponseWriter, r *http.Request) {
	listID, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err := s.DB.UncheckAll(listID); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	items, err := s.DB.GetItems(listID)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	list, err := s.DB.GetList(listID)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	// Return refreshed item list; stats refresh via htmx afterSwap
	s.render(w, "item-list.html", map[string]any{
		"T":     func(k string) string { return s.t(r, k) },
		"Lang":  s.lang(r),
		"List":  list,
		"Items": items,
	})
}