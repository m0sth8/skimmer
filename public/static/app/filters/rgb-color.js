'use strict';
angular.module('skimmerApp')
    .filter('getRgbColor', function() {
        return function(color) {
            if (color){
                return {"color": "rgb(" + color[0] + "," + color[1] + "," + color[2] + ")"};
            } else {
                return {}
            }
        };
    });