package middlewares

import "net/http"

func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// disables dns prefetching (a browser resolving a domain name in the background while user is browsing a page) Reduces risk of DNS attacks
		w.Header().Set("X-DNS-Prefetch-Control", "off")
		// prevents webpage from being displayed in an iframe or on another website; helps avoid clickjacking attacks
		w.Header().Set("X-Frame-Options", "DENY")
		// enables the cross-site scripting filter that is usually built into browsers
		w.Header().Set("X-XSS-Protection", "1;mode=block")
		// prevents browsers from mime sniffing and tells the browser instead to just trust the content-type provided by server
		w.Header().Set("X-Content-Type-Options", "nosniff")
		// enforces https for max age in seconds (tells the browser only to connect through https and says to remember this for the specified time)
		w.Header().Set("Strict Transport Security", "max-age=63072000;includeSubDomains;preload")
		// controls which resources can be loaded on the page - in this case stating that only resources from the server are allowed
		w.Header().Set("Content-Security-Policy", "default-src 'self'")
		// controls how much referrer information is included with requests made from your site
		w.Header().Set("Referrer-Policy", "no-referrer")
		//hides the backend technologies used so that the site is more difficult to hack; deliberately incorrect header
		w.Header().Set("X-Powered-By", "rubyonrails")
		next.ServeHTTP(w, r)
	})
}
