'use strict';
angular.module('skimmerApp')
    .filter('humanSize', function() {
        return function(size) {
            if (size > 1024 * 1024 * 2) {
                return (size / 1024 / 1024).toFixed(2) + " Mb"
            }
            if (size > 1024 * 10) {
                return (size / 1024).toFixed(2) + " Kb"
            }
            return size + " bytes"
        };
    });




