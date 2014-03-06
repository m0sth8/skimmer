'use strict';

angular.module('skimmerApp')
    .factory('skBackendService', function($timeout, $q, $filter, Restangular) {

        var createBin = function(priv){
            return Restangular
                .all('bins/')
                .post({"private": priv})
        };

        var updateBin = function(bin){
            return Restangular
                .all('bins')
                .one(bin.name)
                .customPUT(bin)
        };

        var getHistory = function(){
            return Restangular
                .one('bins/')
                .getList()
        };

        var getRequests = function(qBin){
            var qDefer = $q.defer();
            qBin.then(function(bin){
                Restangular
                    .one('bins', bin.name)
                    .getList('requests/')
                    .then(qDefer.resolve);
            });
            return qDefer.promise;

        };

        var getRequest = function(qBin, id){
            var qDefer = $q.defer();
            qBin.then(function(bin){
                Restangular
                    .one('bins', bin.name)
                    .one('requests', id)
                    .get()
                    .then(qDefer.resolve, qDefer.reject);
            });
            return qDefer.promise;

        };

        var getBin = function(name){
            return Restangular
                .one('bins', name)
                .get()
        };

        return {
            'createBin': createBin,
            'getHistory': getHistory,
            'getRequests': getRequests,
            'getRequest': getRequest,
            'getBin': getBin,
            'updateBin': updateBin
        };
    });

