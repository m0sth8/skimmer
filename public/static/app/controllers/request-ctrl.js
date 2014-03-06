'use strict';
angular.module('skimmerApp')
    .controller('RequestCtrl', function($scope, $stateParams, $state, $location, $rootScope,
                                    skBackendService) {

        var qBin = skBackendService.getBin($stateParams.binName);
        var qRequest = skBackendService.getRequest(qBin, $stateParams.requestId)

        $scope.bin = qBin.$object;
        $scope.binColor = "";
        $scope.request = null;

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

        qRequest.then(function(request){
            $scope.request = request;
        }, function(err){
            if (err.data && err.data.error){
                $scope.error = err.data.error;
            }
        });

        $scope.getBinUrl = function(bin) {
            return location.protocol + "//" + location.host + "/bins/" + bin.name;
        }
        $scope.getRequestUrl = function(bin, request) {
            return location.protocol + "//" + location.host + "/bins/" + bin.name + "/requests/" + request.id;
        }
    });
