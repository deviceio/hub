// core imports
import { BrowserModule } from '@angular/platform-browser';
import { NgModule } from '@angular/core';
import { RouterModule, Routes } from '@angular/router';

// app imports
import { AppComponent } from './app.component';
import { AppShellComponent } from './app-shell.component';
import { HubDashboardComponent } from './dashboard/hub-dashboard.component';
import { HubDevicesComponent } from './devices/hub-devices.component';
import { PageNotFoundComponent } from './shared/page-not-found.component';
import { LoginFormComponent } from './login/login-form.component';
import { SessionService } from './_services/session.service';
import { AuthGuard } from './_guards/auth.guard';

// angular material imports
import { BrowserAnimationsModule } from '@angular/platform-browser/animations';
import { MdSidenavModule } from '@angular/material';
import { MdToolbarModule } from '@angular/material';
import { MdButtonModule } from '@angular/material';
import { MdCardModule } from '@angular/material';
import { MdProgressSpinnerModule } from '@angular/material';
import { MdInputModule } from '@angular/material';

const appRoutes: Routes = [
    {
        path: '',
        component: AppShellComponent,
        canActivate: [AuthGuard],
        children: [
            {
                path: '',
                component: HubDashboardComponent
            },
            {
                path: 'dashboard',
                component: HubDashboardComponent
            },
            {
                path: 'devices',
                component: HubDevicesComponent
            },
        ]
    },
    {
        path: 'login',
        component: LoginFormComponent
    },
    {
        path: '**',
        component: PageNotFoundComponent
    }
];

@NgModule({
    declarations: [
        AppComponent,
        AppShellComponent,
        HubDashboardComponent,
        HubDevicesComponent,
        PageNotFoundComponent,
        LoginFormComponent
    ],
    imports: [
        BrowserModule,
        BrowserAnimationsModule,
        MdSidenavModule,
        MdToolbarModule,
        MdButtonModule,
        MdCardModule,
        MdProgressSpinnerModule,
        MdInputModule,
        RouterModule.forRoot(appRoutes, {
            enableTracing: true
        })
    ],
    providers: [
        SessionService,
        AuthGuard
    ],
    bootstrap: [AppComponent]
})
export class AppModule { }
