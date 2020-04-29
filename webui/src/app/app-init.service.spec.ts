import { TestBed } from '@angular/core/testing'

import { AppInitService } from './app-init.service'
import { UsersService } from './backend'
import { HttpClient, HttpHandler } from '@angular/common/http'

describe('AppInitService', () => {
    beforeEach(() =>
        TestBed.configureTestingModule({
            providers: [UsersService, HttpClient, HttpHandler],
        })
    )

    it('should be created', () => {
        const service: AppInitService = TestBed.inject(AppInitService)
        expect(service).toBeTruthy()
    })
})
