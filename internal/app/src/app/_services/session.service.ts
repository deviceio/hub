import { Injectable } from '@angular/core';
import { Router } from '@angular/router';

@Injectable()
export class SessionService {
    private _authenticated = false;

    constructor(private router: Router) {
        this._authenticated = sessionStorage.getItem('authenticated') === 'true' ? true : false;
    }

    public get authenticated(): boolean {
        return this._authenticated;
    }

    public async create(username: string, password: string): Promise<boolean> {
        return new Promise<boolean>((resolve, reject) => {
            setTimeout(() => {
                console.log('logging in');
                if (username === 'admin' && password === 'admin') {
                    this._authenticated = true;
                    sessionStorage.setItem('authenticated', 'true');
                    resolve(true);
                } else {
                    resolve(false);
                }
            }, 1000);
        });
    }

    public async destroy() {
        this._authenticated = false;
        sessionStorage.removeItem('authenticated');
        this.router.navigate(['/login'], {});
    }
}
