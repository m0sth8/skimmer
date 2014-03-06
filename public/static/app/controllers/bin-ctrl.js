'use strict';
angular.module('skimmerApp')
    .controller('BinCtrl', function($scope, $stateParams, $state, $location, $rootScope,
                                    skBackendService) {

        var qBin = skBackendService.getBin($stateParams.binName);
        var qRequests = skBackendService.getRequests(qBin)

        $scope.bin = qBin.$object;
        $scope.requests = [];
        $scope.binColor = "";
        $scope.error = null;


        qBin.then(function(bin){
            $rootScope.title = bin.name + " bin";
            if (bin.favicon != ""){
                $rootScope.faviconHref = bin.favicon;
            }

            $scope.binColor = "#" + Number(bin.color[0]).toString(16)
                         + Number(bin.color[1]).toString(16)
                         + Number(bin.color[2]).toString(16);

        }, function(err){
            if (err.data && err.data.error){
                $scope.error = err.data.error;
            }
        });

        qRequests.then(function(requests){
            $scope.requests = requests;
        });

        $scope.getBinUrl = function(bin) {
            return location.protocol + "//" + location.host + "/bins/" + bin.name;
        };

        $scope.update = function() {
            skBackendService.updateBin($scope.bin);
        };
    });