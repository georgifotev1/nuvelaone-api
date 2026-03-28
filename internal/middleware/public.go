package middleware

// // 1. Always runs on every /p/{slug} route — resolves tenant into context
// func TenantResolver(tenantService domain.TenantService) func(http.Handler) http.Handler {
//     return func(next http.Handler) http.Handler {
//         return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//             slug := chi.URLParam(r, "slug")
//             if slug == "" {
//                 jsonutil.WriteError(w, http.StatusBadRequest, "missing tenant")
//                 return
//             }
//             tenant, err := tenantService.GetBySlug(r.Context(), slug)
//             if err != nil {
//                 jsonutil.WriteError(w, http.StatusNotFound, "tenant not found")
//                 return
//             }
//             ctx := domain.ContextWithTenant(r.Context(), tenant)
//             next.ServeHTTP(w, r.WithContext(ctx))
//         })
//     }
// }
//
// // 2. Only on protected customer routes — requires a valid customer JWT
// func CustomerJWTAuth(secret string) func(http.Handler) http.Handler {
//     return func(next http.Handler) http.Handler {
//         return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//             token := extractToken(r)
//             if token == "" {
//                 jsonutil.WriteError(w, http.StatusUnauthorized, "unauthorized")
//                 return
//             }
//             claims, err := auth.ParseCustomerToken(token, secret)
//             if err != nil {
//                 jsonutil.WriteError(w, http.StatusUnauthorized, "unauthorized")
//                 return
//             }
//             // Sanity check: token's tenant must match the resolved tenant in context
//             tenant := domain.TenantFromContext(r.Context())
//             if tenant == nil || claims.TenantID != tenant.ID {
//                 jsonutil.WriteError(w, http.StatusForbidden, "forbidden")
//                 return
//             }
//             ctx := auth.ContextWithCustomerClaims(r.Context(), claims)
//             next.ServeHTTP(w, r.WithContext(ctx))
//         })
//     }
// }

// r.Route("/p/{slug}", func(r chi.Router) {
//     r.Use(middleware.TenantResolver(app.tenantService)) // always runs first
//
//     r.Post("/auth/register", customerAuthHandler.Register) // tenant in ctx, no customer auth
//     r.Post("/auth/login",    customerAuthHandler.Login)
//
//     r.Get("/services",       serviceHandler.ListPublic)    // tenant in ctx, no customer auth
//     r.Get("/timeslots",      appointmentHandler.Timeslots) // same
//
//     r.Group(func(r chi.Router) {
//         r.Use(middleware.CustomerJWTAuth(app.config.Auth.CustomerJWTSecret))
//         r.Get("/me",              customerHandler.GetMe)
//         r.Post("/appointments",   appointmentHandler.Create)
//         r.Get("/appointments",    appointmentHandler.ListMine)
//     })
// })
