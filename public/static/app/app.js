'use strict';
angular.module('skimmerApp', ['restangular','ui.router'])
    .config(function($locationProvider, $httpProvider, $stateProvider, $urlRouterProvider, RestangularProvider) {
        $locationProvider.html5Mode(true);
        $httpProvider.defaults.headers.common['Content-Type'] = 'application/json';

        $stateProvider
            .state('home', {
                "url": "/",
                "views": {
                    "ViewMain": {
                        "templateUrl": "/static/views/home.html",
                        "controller": "HomeCtrl"
                    }
                }
            })
            .state('request', {
                "url": "/inspect/:binName/requests/:requestId",
                "views": {
                    "ViewMain": {
                        "templateUrl": "/static/views/request.html",
                        "controller": "RequestCtrl"
                    },
                    "ViewHead": {
                        "templateUrl": "/static/views/bin-url.html",
                        "controller": "BinCtrl"
                    }
                }
            })
            .state('bin', {
                "url": "/inspect/:binName",
                "views": {
                    "ViewMain": {
                        "templateUrl": "/static/views/bin.html",
                        "controller": "BinCtrl"
                    },
                    "ViewHead": {
                        "templateUrl": "/static/views/bin-url.html",
                        "controller": "BinCtrl"
                    }
                }
            });

        RestangularProvider.setBaseUrl('/api/v1');

    });