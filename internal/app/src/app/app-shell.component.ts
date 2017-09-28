import { Component } from '@angular/core';
import { Routes } from '@angular/router';
import { SessionService } from './_services/session.service';

@Component({
  selector: 'app-shell',
  templateUrl: './app-shell.component.html',
  styleUrls: ['./app-shell.component.css']
})
export class AppShellComponent {
    constructor(public session: SessionService) {
    }
}
