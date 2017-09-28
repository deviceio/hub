import { Component } from '@angular/core';
import { SessionService } from '../_services/session.service';
import { Router} from '@angular/router';

@Component({
  selector: 'login-form',
  templateUrl: './login-form.component.html',
  styleUrls: ['./login-form.component.css']
})
export class LoginFormComponent {
    public username = '';
    public password = '';
    public error: string;
    public loading = false;

    constructor(public session: SessionService, private router: Router) {
    }

    public async login() {
        this.loading = true;
        this.error = '';
        const success = await this.session.create(this.username, this.password);

        if (!success) {
            this.error = 'invalid username or password';
        }

        this.loading = false;

        if (success) {
            this.username = '';
            this.password = '';
            this.router.navigate(['/']);
        }
    }
}
