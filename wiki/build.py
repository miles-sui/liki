"""Build trilingual naming wiki from content/**/*.md — Liki v4.0.

Source .md format:
  YAML: slug, hub, type, refs
  Body: # Title, ## Direct Answer, ## Key Signals, ## Decision Map, ## Why You're Asking

Output:
  web/wiki/
  ├── index.html            (redirect → zh-hant/index.html)
  ├── style.css
  ├── sitemap.xml
  ├── robots.txt
  ├── zh-hant/
  ├── zh-hans/
  └── en/
"""
import os, re, json
import zhconv

ROOT = os.path.dirname(os.path.abspath(__file__))
CONTENT = os.path.join(ROOT, "content")
OUT = os.path.join(ROOT, "..", "web", "wiki")

def s2t(text): return zhconv.convert(text, "zh-hant")

LANG = {
    "zh-hant": {"html": "zh-Hant", "label": "繁體中文", "dir": "zh-hant/"},
    "zh-hans": {"html": "zh-Hans", "label": "简体中文", "dir": "zh-hans/"},
    "en":      {"html": "en",      "label": "English",  "dir": "en/"},
}

def depth(path): return path.count("/")
def rel_css(slug): return ("../" * (depth(slug) + 1)) + "style.css"

def parse_md(path):
    with open(path) as f: text = f.read()
    m = re.match(r'^---\n(.*?)\n---\n(.*)', text, re.DOTALL)
    if not m: raise ValueError(f"Invalid frontmatter in {path}")
    fm, body = m.group(1), m.group(2).strip()
    data, ck = {}, None
    for line in fm.split("\n"):
        line = line.rstrip()
        if not line: continue
        if ":" in line and not line.startswith("  "):
            key, _, val = line.partition(":"); key = key.strip(); val = val.strip()
            data[key] = [val] if val else []; ck = key
        elif line.startswith("  - "):
            if ck and ck in data: data[ck].append(line[4:])
    title_m = re.match(r'^#\s+(.*)', body.split("\n")[0])
    title = title_m.group(1) if title_m else ""
    lines = body.split("\n")
    title_en = ""
    for li in range(1, min(5, len(lines))):
        m_en = re.match(r'^#\s+(.*)\s+\(EN\)', lines[li])
        if m_en: title_en = m_en.group(1); break
    sections, cs = {}, None
    for line in lines:
        m2 = re.match(r'^##\s+(.*)', line)
        if m2: cs = m2.group(1).strip(); sections[cs] = []
        elif cs: sections[cs].append(line)
    for k in sections: sections[k] = "\n".join(sections[k]).strip()
    sv = data.get("slug", [""]); hv = data.get("hub", [""])
    return {
        "slug": sv[0] if isinstance(sv, list) and sv else (sv if isinstance(sv, str) else ""),
        "hub": hv[0] if isinstance(hv, list) and hv else (hv if isinstance(hv, str) else ""),
        "title": title, "title_en": title_en, "sections": sections,
    }

def load_entries():
    es = []
    for root, dirs, files in os.walk(CONTENT):
        for f in sorted(files):
            if f.endswith(".md"): es.append(parse_md(os.path.join(root, f)))
    return es

def build_page(entry, lang):
    cfg = LANG[lang]; slug, title = entry["slug"], entry["title"]
    title_disp = s2t(title) if lang == "zh-hant" else (entry.get("title_en") or title) if lang == "en" else title
    sec = entry["sections"]

    if lang == "en":
        labels = {"Direct Answer": "Direct Answer",
            "Key Signals · BaZi 指标": "Key Signals · BaZi",
            "Decision Map": "Decision Map", "Why You're Asking": "Why You're Asking"}
        pg = "— Liki AI Naming Advisor"
        for key in list(sec.keys()):
            if key.endswith(" (EN)"):
                base = key[:-5]
                if base in sec: sec[base] = sec[key]
    elif lang == "zh-hant":
        labels = {"Direct Answer": "直接答案",
            "Key Signals · BaZi 指标": "關鍵信號 · BaZi",
            "Decision Map": "決策地圖", "Why You're Asking": "你為什麼會問"}
        pg = "— 靈機 Liki · AI起名顧問"
    else:
        labels = {"Direct Answer": "直接答案",
            "Key Signals · BaZi 指标": "关键信号 · BaZi",
            "Decision Map": "决策地图", "Why You're Asking": "你为什么会问"}
        pg = "— 灵机 Liki · AI起名顾问"

    page_title = f"{title_disp} {pg}"
    da = sec.get("Direct Answer", "")
    desc = (s2t(da) if lang == "zh-hant" else da)[:160]
    canonical = f"https://liki.hk/wiki/{cfg['dir']}{slug}.html"
    css_rel = rel_css(slug)

    ld = json.dumps({
        "@context": "https://schema.org",
        "@type": "FAQPage",
        "mainEntity": [{
            "@type": "Question",
            "name": title_disp,
            "acceptedAnswer": {"@type": "Answer", "text": da.replace('\n', ' ').strip()}
        }]
    }, ensure_ascii=False)

    p = ['<!DOCTYPE html>', f'<html lang="{cfg["html"]}">', '<head>',
         '<meta charset="UTF-8">',
         '<meta name="viewport" content="width=device-width, initial-scale=1">',
         f'<title>{page_title}</title>', f'<meta name="description" content="{desc}">',
         f'<link rel="canonical" href="{canonical}">',
         f'<link rel="alternate" hreflang="zh-hant" href="https://liki.hk/wiki/zh-hant/{slug}.html">',
         f'<link rel="alternate" hreflang="zh-hans" href="https://liki.hk/wiki/zh-hans/{slug}.html">',
         f'<link rel="alternate" hreflang="en" href="https://liki.hk/wiki/en/{slug}.html">',
         '<link rel="alternate" hreflang="x-default" href="https://liki.hk/wiki/zh-hant/index.html">',
         f'<link rel="stylesheet" href="{css_rel}">',
         f'<script type="application/ld+json">{ld}</script>',
         '</head>', '<body>',
         f'<nav><a href="{("../" * (depth(slug) + 1))}index.html">{s2t("← 首頁") if lang == "zh-hant" else ("← Home" if lang == "en" else "← 首页")}</a></nav>',
         '<main>', f'<h1>{title_disp}</h1>',
         '<div class="lang-switch">']
    for code in ["zh-hant", "zh-hans"]:
        lc = LANG[code]
        if code == lang: p.append(f'<strong>{lc["label"]}</strong>')
        else: p.append(f'<a href="{("../" * (depth(slug) + 1))}{LANG[code]["dir"]}{slug}.html">{lc["label"]}</a>')
    p.append('</div>')

    order = ["Direct Answer"]
    order += [sk for sk in sorted(sec.keys()) if sk.startswith("Key Signals") and not sk.endswith(" (EN)")]
    order += ["Decision Map", "Why You're Asking"]
    for sk in order:
        if sk not in sec: continue
        c = sec[sk]; lab = labels.get(sk, sk)
        if lang == "zh-hant": lab = s2t(lab); c = s2t(c)
        p.append('<section>'); p.append(f'<h2>{lab}</h2>')
        if sk.startswith("Key Signals") or sk == "Decision Map":
            items = [l[2:].strip() if l.startswith("- ") else l[1:].strip() if l.startswith("-") else l.strip()
                     for l in c.split("\n") if l.strip().startswith("-")]
            if items:
                p.append('<ul>')
                for it in items: p.append(f'<li>{s2t(it) if lang == "zh-hant" else it}</li>')
                p.append('</ul>')
            else: p.append(f'<p>{c}</p>')
        else:
            for para in c.split("\n\n"):
                para = para.strip()
                if para: p.append(f'<p>{para}</p>')
        p.append('</section>')

    # CTA
    if lang == "zh-hant":
        p.append('<div style="margin-top:2rem;padding:1rem;background:#f8f8f8;border-radius:4px;font-size:0.9rem">')
        p.append('<strong>想為你的孩子起一個有意義的名字？</strong><br>')
        p.append('在靈機獲取你的專屬起名建議 → <a href="https://liki.hk">liki.hk</a></div>')
    elif lang == "en":
        p.append('<div style="margin-top:2rem;padding:1rem;background:#f8f8f8;border-radius:4px;font-size:0.9rem">')
        p.append('<strong>Looking for a meaningful Chinese name?</strong><br>')
        p.append('Get personalized naming advice at Liki → <a href="https://liki.hk">liki.hk</a></div>')
    else:
        p.append('<div style="margin-top:2rem;padding:1rem;background:#f8f8f8;border-radius:4px;font-size:0.9rem">')
        p.append('<strong>想为你的孩子起一个有意义的名字？</strong><br>')
        p.append('在灵机获取你的专属起名建议 → <a href="https://liki.hk">liki.hk</a></div>')
    p.append('</main>'); p.append('</body>'); p.append('</html>')
    return '\n'.join(p) + '\n'

def build_index(lang, entries):
    cfg = LANG[lang]
    if lang == "en":
        pt = "Liki — AI Naming Advisor"; hd = "Liki Naming Wiki"
        slogan = "Meaningful Names Start Here."
        caps = "AI-powered Chinese naming guidance"
    elif lang == "zh-hant":
        pt = "靈機 Liki · 起名知識庫"; hd = "靈機 Liki · 起名知識庫"
        slogan = "名正言順。"; caps = "有意義的名字，始於理解。"
    else:
        pt = "灵机 Liki · 起名知识库"; hd = "灵机 Liki · 起名知识库"
        slogan = "名正言顺。"; caps = "有意义的名字，始于理解。"
    canonical = f"https://liki.hk/wiki/{cfg['dir']}index.html"
    css_rel = "../style.css"

    lines = ['<!DOCTYPE html>', f'<html lang="{cfg["html"]}">', '<head>',
             '<meta charset="UTF-8">', '<meta name="viewport" content="width=device-width, initial-scale=1">',
             f'<title>{pt}</title>', f'<meta name="description" content="{caps}">',
             f'<link rel="canonical" href="{canonical}">',
             '<link rel="alternate" hreflang="zh-hant" href="https://liki.hk/wiki/zh-hant/index.html">',
             '<link rel="alternate" hreflang="zh-hans" href="https://liki.hk/wiki/zh-hans/index.html">',
             '<link rel="alternate" hreflang="x-default" href="https://liki.hk/wiki/zh-hant/index.html">',
             f'<link rel="stylesheet" href="{css_rel}">', '</head>', '<body>', '<main>',
             f'<h1>{hd}</h1>',
             f'<p style="font-size:1.1rem;color:#555;margin-top:-0.5rem">{slogan}</p>',
             f'<p style="font-size:0.9rem;color:#888">{caps}</p>',
             '<div class="lang-switch">']
    for code in ["zh-hant", "zh-hans"]:
        lc = LANG[code]
        if code == lang: lines.append(f'<strong>{lc["label"]}</strong>')
        else: lines.append(f'<a href="../{LANG[code]["dir"]}index.html">{lc["label"]}</a>')
    lines.append('</div>')

    lines.append('<section><h2>起名知识</h2><ul>')
    for e in entries:
        tl = s2t(e["title"]) if lang == "zh-hant" else e["title"]
        lines.append(f'<li><a href="{e["slug"]}.html">{tl}</a></li>')
    lines.append('</ul></section>')
    lines.append('</main>'); lines.append('</body>'); lines.append('</html>')
    return '\n'.join(lines) + '\n'

build_redirect = lambda: """<!DOCTYPE html>
<html lang="zh-Hant"><head><meta charset="UTF-8">
<meta http-equiv="refresh" content="0; url=zh-hant/index.html">
<link rel="canonical" href="https://liki.hk/wiki/zh-hant/index.html">
<link rel="alternate" hreflang="zh-hant" href="https://liki.hk/wiki/zh-hant/index.html">
<link rel="alternate" hreflang="zh-hans" href="https://liki.hk/wiki/zh-hans/index.html">
<link rel="alternate" hreflang="x-default" href="https://liki.hk/wiki/zh-hant/index.html">
<title>靈機 Liki · 起名知識庫</title></head>
<body><p><a href="zh-hant/index.html">繁體中文</a> |
<a href="zh-hans/index.html">简体中文</a></p></body></html>"""

build_css = lambda: """body{max-width:720px;margin:0 auto;padding:2rem 1.5rem;font-family:system-ui,-apple-system,sans-serif;line-height:1.7;color:#1a1a1a;background:#fff}
nav{margin-bottom:1.2rem;font-size:.85rem;color:#888}
nav a{color:#555;text-decoration:none}
.lang-switch{margin-bottom:1.5rem;font-size:.85rem;color:#888}
.lang-switch a,.lang-switch strong{margin-right:.8rem}
.lang-switch a{color:#0366d6}
h1{font-size:1.8rem;margin-bottom:.5rem}
h2{font-size:1.2rem;margin-top:2rem;border-bottom:1px solid #eee;padding-bottom:.3rem}
section{margin-bottom:1.5rem}
ul{padding-left:1.5rem}
li{margin-bottom:.35rem}
a{color:#0366d6}
@media(prefers-color-scheme:dark){body{color:#ddd;background:#1a1a1a}h2{border-color:#333}a{color:#58a6ff}.lang-switch a{color:#58a6ff}}
"""

def main():
    entries = load_entries()
    entries.sort(key=lambda x: x["slug"])

    # Compute related questions (circular)
    n = len(entries)
    for i, e in enumerate(entries):
        e["_related"] = []
        for j in range(1, min(3, n)):
            r = entries[(i + j) % n]
            e["_related"].append({"title": r["title"], "title_en": r.get("title_en", ""), "slug": r["slug"]})

    for d in ["zh-hant", "zh-hans"]:
        os.makedirs(os.path.join(OUT, d), exist_ok=True)
    os.makedirs(os.path.join(OUT, "en"), exist_ok=True)

    with open(os.path.join(OUT, "style.css"), "w") as f: f.write(build_css())
    with open(os.path.join(OUT, "index.html"), "w") as f: f.write(build_redirect())

    for lang in ["zh-hant", "zh-hans", "en"]:
        lo = os.path.join(OUT, LANG[lang]["dir"])
        with open(os.path.join(lo, "index.html"), "w") as f: f.write(build_index(lang, entries))
        for e in entries:
            sd = os.path.dirname(e["slug"])
            if sd: os.makedirs(os.path.join(lo, sd), exist_ok=True)
        for e in entries:
            with open(os.path.join(lo, f'{e["slug"]}.html'), "w") as f: f.write(build_page(e, lang))

    # Sitemap
    base = "https://liki.hk/wiki"
    sm = ['<?xml version="1.0" encoding="UTF-8"?>',
          '<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9"',
          '        xmlns:xhtml="http://www.w3.org/1999/xhtml">']
    for e in entries:
        slug = e["slug"]
        sm.append(f"<url><loc>{base}/zh-hant/{slug}.html</loc>")
        sm.append(f'<xhtml:link rel="alternate" hreflang="zh-hant" href="{base}/zh-hant/{slug}.html"/>')
        sm.append(f'<xhtml:link rel="alternate" hreflang="zh-hans" href="{base}/zh-hans/{slug}.html"/>')
        sm.append("</url>")
    for lang in ["zh-hant", "zh-hans"]:
        sm.append(f"<url><loc>{base}/{lang}/index.html</loc></url>")
    sm.append("</urlset>")
    with open(os.path.join(OUT, "sitemap.xml"), "w") as f: f.write("\n".join(sm) + "\n")

    with open(os.path.join(OUT, "robots.txt"), "w") as f:
        f.write("Sitemap: https://liki.hk/wiki/sitemap.xml\n\nUser-agent: *\nAllow: /\n")

    print(f"Liki Naming Wiki v4.0: {len(entries)} 篇文章")
    print(f"  输出: web/wiki/")

if __name__ == "__main__":
    main()
