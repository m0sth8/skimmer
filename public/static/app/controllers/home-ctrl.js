'use strict';
angular.module('skimmerApp')
    .controller('HomeCtrl', function($scope, $stateParams, $state, $location, $rootScope,
                                        skBackendService) {
        $rootScope.title = "";

        var qHistory = skBackendService.getHistory();

        $scope.history = qHistory.$object;
        $scope.private = false;

        $scope.createBin = function(){
            skBackendService.createBin($scope.private).then(function(bin){
                $state.transitionTo('bin', {'binName': bin.name});
            })
        };



    });