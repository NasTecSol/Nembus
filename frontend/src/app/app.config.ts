import { ApplicationConfig, provideZoneChangeDetection, importProvidersFrom,} from "@angular/core";
import { provideAnimations } from "@angular/platform-browser/animations";
import { provideToastr } from "ngx-toastr";
import { provideRouter } from "@angular/router";
import { provideHttpClient, HTTP_INTERCEPTORS,withInterceptorsFromDi} from "@angular/common/http";
import { routes } from "./app.routes";
import { TranslateModule, TranslateLoader } from "@ngx-translate/core";
import { HttpClient } from "@angular/common/http";
import { HttpLoaderFactory } from "./utils/translate-loader";
import { AuthInterceptor } from "./core/interceptors/auth.interceptor";
import { LoaderInterceptor } from "./core/interceptors/loader.interceptor";
import { TenantInterceptor } from "./core/interceptors/tenant.interceptor";
import { NgxsStoragePluginModule, StorageOption } from "@ngxs/storage-plugin";
import { NgxsModule, provideStore } from "@ngxs/store";
import { CartState } from "./core/store/cart/cart.state";

export const appConfig: ApplicationConfig = {
  providers: [
    provideZoneChangeDetection({ eventCoalescing: true }),
    provideAnimations(),
    provideToastr({
      timeOut: 3000,
      positionClass: "toast-bottom-right",
      preventDuplicates: true,
      closeButton: true,
    }),
    importProvidersFrom(
      NgxsModule.forRoot([CartState], {
        developmentMode: true,
      }),
       // Persist cart state so it survives page refreshes within the same session
      NgxsStoragePluginModule.forRoot({
        keys: ['cart'],
        storage: StorageOption.SessionStorage, // clears when browser tab closes
        migrations: [
          {
            // Bump this version number whenever CartStateModel gains new required fields.
            // Any persisted state without a matching version gets migrated (defaults merged in).
            version: 2,
            versionKey: '__cartSchemaVersion',
            migrate: (state: any) => ({
              taxAmount: 0,      // new required field in v2
              ...state,
              __cartSchemaVersion: 2,
            }),
          },
        ],
      })
    ),
    provideHttpClient(withInterceptorsFromDi()),
    provideRouter(routes),
    {
      provide: HTTP_INTERCEPTORS,
      useClass: LoaderInterceptor,
      multi: true,
    },
    {
      provide: HTTP_INTERCEPTORS,
      useClass: AuthInterceptor,
      multi: true,
    },
    {
      provide: HTTP_INTERCEPTORS,
      useClass: TenantInterceptor,
      multi: true,
    },
    importProvidersFrom(
      TranslateModule.forRoot({
        defaultLanguage: "en",
        loader: {
          provide: TranslateLoader,
          useFactory: HttpLoaderFactory,
          deps: [HttpClient],
        },
      })
    ),
  ],
};