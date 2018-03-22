/*
  Copyright 2014 Google Inc. All rights reserved.

  Licensed under the Apache License, Version 2.0 (the "License");
  you may not use this file except in compliance with the License.
  You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

  Unless required by applicable law or agreed to in writing, software
  distributed under the License is distributed on an "AS IS" BASIS,
  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
  See the License for the specific language governing permissions and
  limitations under the License.
*/

goog.provide('cardcpx.allbox');
goog.provide('cardcpx.allbox.module');

/**
 * Directive function for ccxAllbox.
 * @param {!angular.$parse} $parse
 * @ngInject
 */
cardcpx.allbox.directive = function($parse) {
  return {
      restrict: 'E',
      template: '<input type="checkbox">',
      replace: true,
      link: function(scope, element, attrs) {
        var check = $parse(attrs['check']);
        var uncheck = $parse(attrs['uncheck']);
        var anyExpr = attrs['any'];
        var allExpr = attrs['all'];
        var any, all;

        var refresh = function() {
          if (all) {
            element[0].checked = true;
            element[0].indeterminate = false;
          } else if (any) {
            element[0].checked = true;
            element[0].indeterminate = true;
          } else {
            element[0].checked = false;
            element[0].indeterminate = false;
          }
        };
        scope.$watch(anyExpr, function(val) {
          any = val;
          refresh();
        });
        scope.$watch(allExpr, function(val) {
          all = val;
          refresh();
        });
        element.on('click', function(event) {
          scope.$apply(function() {
            if (all) {
              uncheck(scope, {$event: event});
            } else {
              check(scope, {$event: event});
            }
          });
        });
      }
  };
};

/** @type {!angular.Module} */
cardcpx.allbox.module = angular.module('cardcpx.allbox', []).
    directive('ccxAllbox', cardcpx.allbox.directive);
