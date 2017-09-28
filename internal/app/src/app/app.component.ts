import { Component, OnInit } from '@angular/core';
import { SessionService } from './_services/session.service';

@Component({
  selector: 'app-root',
  templateUrl: './app.component.html',
  styleUrls: ['./app.component.css']
})
export class AppComponent implements OnInit {
    constructor(public session: SessionService) {
    }

    ngOnInit() {
        document.getElementById('preload-message').remove();
    }
}
