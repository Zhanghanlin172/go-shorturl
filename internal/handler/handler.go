package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"go-shorturl/internal/generator"
	"go-shorturl/internal/store"
)

type Handler struct {
	store *store.MemoryStore
}

func New(s *store.MemoryStore) *Handler {
	return &Handler{store: s}
}

type ShortenReq struct {
	URL string `json:"url"`
}

type ShortenResp struct {
	ShortURL string `json:"short_url"`
	Code     string `json:"code"`
	LongURL  string `json:"long_url"`
}

// Shorten POST /api/shorten  { "url": "https://..." }
func (h *Handler) Shorten(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"method not allowed"}`, 405)
		return
	}

	var req ShortenReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.URL == "" {
		http.Error(w, `{"error":"invalid body, need {\"url\":\"...\"}"}`, 400)
		return
	}

	if !strings.HasPrefix(req.URL, "http://") && !strings.HasPrefix(req.URL, "https://") {
		http.Error(w, `{"error":"url must start with http:// or https://"}`, 400)
		return
	}

	code := generator.Generate(req.URL)

	entry := &store.ShortURL{
		Code:      code,
		LongURL:   req.URL,
		CreatedAt: time.Now(),
	}
	h.store.Save(entry)

	log.Printf("[SHORTEN] %s -> %s", code, req.URL)

	resp := ShortenResp{
		ShortURL: fmt.Sprintf("http://%s/%s", r.Host, code),
		Code:     code,
		LongURL:  req.URL,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// Stats GET /api/stats/{code}
func (h *Handler) Stats(w http.ResponseWriter, r *http.Request) {
	code := strings.TrimPrefix(r.URL.Path, "/api/stats/")
	if code == "" {
		http.Error(w, `{"error":"missing code"}`, 400)
		return
	}

	entry, ok := h.store.Find(code)
	if !ok {
		http.Error(w, `{"error":"not found"}`, 404)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(entry)
}

// Health GET /api/health
func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"status":"ok","stored_urls":%d}`, h.store.Count())
}

// Redirect GET /{code} -> 301 redirect
func (h *Handler) Redirect(w http.ResponseWriter, r *http.Request) {
	code := strings.TrimPrefix(r.URL.Path, "/")
	if code == "" || code == "favicon.ico" {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(indexPage))
		return
	}

	entry, ok := h.store.Find(code)
	if !ok {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(404)
		w.Write([]byte("<html><body style='font-family:sans-serif;text-align:center;padding-top:80px'><h1>404</h1><p>短链接不存在</p><p><a href='/'>← 返回</a></p></body></html>"))
		return
	}

	h.store.IncrementAccess(code)
	log.Printf("[REDIRECT] %s -> %s (访问第%d次)", code, entry.LongURL, entry.AccessCount+1)

	http.Redirect(w, r, entry.LongURL, http.StatusMovedPermanently)
}

const indexPage = `<!DOCTYPE html>
<html lang="zh-CN">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>短链接服务</title>
<style>
*{margin:0;padding:0;box-sizing:border-box}
body{font-family:-apple-system,sans-serif;background:#f5f6fa;display:flex;justify-content:center;align-items:center;min-height:100vh}
.card{background:#fff;padding:40px;border-radius:16px;box-shadow:0 2px 16px rgba(0,0,0,.06);max-width:480px;width:100%}
h1{font-size:24px;color:#1e293b;margin-bottom:4px}
p{color:#94a3b8;font-size:14px;margin-bottom:24px}
input{width:100%;padding:12px 16px;border:2px solid #e2e8f0;border-radius:8px;font-size:15px;outline:none;margin-bottom:12px}
input:focus{border-color:#2563eb}
button{width:100%;padding:12px;background:#2563eb;color:#fff;border:none;border-radius:8px;font-size:15px;font-weight:600;cursor:pointer}
button:hover{opacity:.9}
.result{margin-top:16px;padding:14px;background:#f0fdf4;border-radius:8px;display:none;word-break:break-all}
.result a{color:#2563eb;font-weight:600;text-decoration:none}
.error{margin-top:12px;color:#ef4444;font-size:13px;display:none}
</style>
</head>
<body>
<div class="card">
<h1>🔗 短链接生成器</h1>
<p>基于 Go 实现的高并发短链接服务</p>
<input id="urlInput" placeholder="粘贴长网址 (https://...)" autofocus>
<button onclick="shorten()">生成短链接</button>
<div class="result" id="result">
  短链接：<a id="shortLink" href="#" target="_blank"></a><br>
  <span style="font-size:12px;color:#666">码：<span id="codeText"></span></span>
</div>
<div class="error" id="error"></div>
</div>
<script>
async function shorten() {
  const url = document.getElementById('urlInput').value.trim();
  const err = document.getElementById('error');
  const res = document.getElementById('result');
  err.style.display = 'none'; res.style.display = 'none';
  if (!url) { err.textContent = '请输入网址'; err.style.display = 'block'; return }
  try {
    const r = await fetch('/api/shorten', { method:'POST', headers:{'Content-Type':'application/json'}, body:JSON.stringify({url}) });
    const d = await r.json();
    if (!r.ok) { err.textContent = d.error; err.style.display = 'block'; return }
    document.getElementById('shortLink').href = d.short_url;
    document.getElementById('shortLink').textContent = d.short_url;
    document.getElementById('codeText').textContent = d.code;
    res.style.display = 'block';
  } catch(e) {
    err.textContent = '请求失败，请检查服务是否启动';
    err.style.display = 'block';
  }
}
document.getElementById('urlInput').addEventListener('keydown', e => { if (e.key==='Enter') shorten() });
</script>
</body>
</html>`
