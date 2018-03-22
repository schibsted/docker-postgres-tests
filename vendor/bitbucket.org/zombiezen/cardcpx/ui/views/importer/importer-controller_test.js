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

goog.require('cardcpx.importer.ImporterCtrl');
goog.require('cardcpx.importer.ImportState');
goog.require('cardcpx.importer.ItemState');
goog.require('cardcpx.importer.module');

describe('ImporterCtrl', function() {
  var clips = [
    {
      name: 'foo',
      paths: ['foo/1.mov', 'foo/2.mov'],
      totalSize: 42
    },
    {
      name: 'bar',
      paths: ['bar.mov'],
      totalSize: 1025
    },
    {
      name: 'already-imported',
      paths: ['casablanca.mov'],
      totalSize: 314159
    }];
  var takes = [
    {
      id: {scene: '2', num: '3a'},
      clipName: 'already-imported',
      select: false
    }];

  beforeEach(module(cardcpx.importer.module.name, function($provide) {
    $provide.service('Importer', function($q) {
      var stat = {enabled: true, active: false};
      this.getStatus = function() {
        var deferred = $q.defer();
        deferred.resolve(stat);
        return deferred.promise;
      };
      this.setStatus = function(s) {
        stat = s;
      };
      this.startImport = function(path, subdir, items) {
        this.path = path;
        this.subdir = subdir;
        this.items = items;

        var deferred = $q.defer();
        deferred.resolve(null);
        return deferred.promise;
      };
      spyOn(this, 'startImport').and.callThrough();
    });
  }));

  var $httpBackend, $timeout;
  beforeEach(inject(function(_$httpBackend_, _$timeout_, $controller) {
    $httpBackend = _$httpBackend_;
    $timeout = _$timeout_;
    $httpBackend.when('GET', '/source?path=%2Ffoo').respond(clips);
    // TODO(light): fake TakeStorage
    $httpBackend.when('GET', '/take/').respond(takes);

    this.newCtrl = function(path) {
      return $controller('ImporterCtrl', {
        '$routeParams': {path: path},
        '$cookies': {}});
    };

    try {
      $timeout.flush();
    } catch(e) {}
  }));
  afterEach(function() {
    $httpBackend.verifyNoOutstandingExpectation();
    $httpBackend.verifyNoOutstandingRequest();
    $timeout.verifyNoPendingTasks();
  });

  it('initializes', function() {
    var ctrl = this.newCtrl('/foo');
    expect(ctrl.path).toEqual('/foo');
    expect(ctrl.enabled).toBe(true);
    expect(ctrl.progress).toBeLessThan(0);
    expect(ctrl.eta).toBeNull();
    expect(ctrl.items).toEqual([]);
    expect(ctrl.getState()).toBe(cardcpx.importer.ImportState.LOADING);
    expect(ctrl.isActive()).toBe(false);
    expect(ctrl.anyChecked()).toBe(false);
    expect(ctrl.allChecked()).toBe(false);

    $httpBackend.flush();
    $timeout.flush();

    expect(ctrl.path).toEqual('/foo');
    expect(ctrl.enabled).toBe(true);
    expect(ctrl.progress).toBeLessThan(0);
    expect(ctrl.eta).toBeNull();
    expect(ctrl.items).toEqual([
      {
        checked: true,
        state: cardcpx.importer.ItemState.READY,
        clip: clips[0],
        scene: 'scratch',
        num: '1',
        select: false
      },
      {
        checked: true,
        state: cardcpx.importer.ItemState.READY,
        clip: clips[1],
        scene: 'scratch',
        num: '2',
        select: false
      },
      {
        checked: false,
        state: cardcpx.importer.ItemState.IMPORTED,
        clip: clips[2],
        scene: 'scratch',
        num: '3',
        select: false
      },
    ]);
    expect(ctrl.getState()).toBe(cardcpx.importer.ImportState.READY);
    expect(ctrl.isActive()).toBe(false);
    expect(ctrl.anyChecked()).toBe(true);
    expect(ctrl.allChecked()).toBe(false);
  });

  it('does not load for empty path', inject(function() {
    var ctrl = this.newCtrl('');
    $timeout.flush();
    expect(ctrl.path).toEqual('');
    expect(ctrl.enabled).toBe(true);
    expect(ctrl.progress).toBeLessThan(0);
    expect(ctrl.eta).toBeNull();
    expect(ctrl.items).toEqual([]);
    expect(ctrl.getState()).toBe(cardcpx.importer.ImportState.NO_DATA);
    expect(ctrl.isActive()).toBe(false);
    expect(ctrl.anyChecked()).toBe(false);
    expect(ctrl.allChecked()).toBe(false);
    $httpBackend.verifyNoOutstandingRequest();
  }));

  it('checks all with checkAll', function() {
    var ctrl = this.newCtrl('/foo');
    $httpBackend.flush();
    $timeout.flush();
    ctrl.checkAll();
    expect(ctrl.items[0].checked).toBe(true);
    expect(ctrl.items[1].checked).toBe(true);
    expect(ctrl.items[2].checked).toBe(true);
  });

  it('unchecks all with checkNone', function() {
    var ctrl = this.newCtrl('/foo');
    $httpBackend.flush();
    $timeout.flush();
    ctrl.checkNone();
    expect(ctrl.items[0].checked).toBe(false);
    expect(ctrl.items[1].checked).toBe(false);
    expect(ctrl.items[2].checked).toBe(false);
  });

  it('autofills scenes by copying to end of list', function() {
    var ctrl = this.newCtrl('/foo');
    $httpBackend.flush();
    $timeout.flush();
    ctrl.items[1].scene = 'blah';
    ctrl.autofillScene(1);
    expect(ctrl.items[0].scene).toEqual('scratch');
    expect(ctrl.items[1].scene).toEqual('blah');
    expect(ctrl.items[2].scene).toEqual('blah');
  });

  it('autofills scenes by incrementing to end of list', function() {
    var ctrl = this.newCtrl('/foo');
    $httpBackend.flush();
    $timeout.flush();
    ctrl.items[1].num = '5';
    ctrl.autofillNum(1);
    expect(ctrl.items[0].num).toEqual('1');
    expect(ctrl.items[1].num).toEqual('5');
    expect(ctrl.items[2].num).toEqual('6');
  });

  it('starts the import', inject(function(Importer) {
    var ctrl = this.newCtrl('/foo');
    $httpBackend.flush();
    $timeout.flush();
    ctrl.items[1].select = true;

    ctrl.startImport();
    expect(Importer.startImport).toHaveBeenCalled();
    expect(Importer.path).toEqual('/foo');
    expect(Importer.subdir).toEqual('');
    expect(Importer.items).toEqual([
      {
        clip: clips[1],
        scene: 'scratch',
        num: '2',
        select: true
      },
      {
        clip: clips[0],
        scene: 'scratch',
        num: '1',
        select: false
      },
    ]);

    // TODO(light): check progress
    $timeout.flush();
  }));

  it('starts the import with a subdirectory', inject(function(Importer) {
    var ctrl = this.newCtrl('/foo');
    $httpBackend.flush();
    $timeout.flush();
    ctrl.subdir = 'SUB';

    ctrl.startImport();

    expect(Importer.startImport).toHaveBeenCalled();
    expect(Importer.subdir).toEqual('SUB');

    // TODO(light): check progress
    $timeout.flush();
  }));
});
