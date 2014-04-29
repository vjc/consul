//
// A Consul service.
//
App.Service = Ember.Object.extend({
  //
  // The number of failing checks within the service.
  //
  failingChecks: function() {
    return this.get('Checks').filterBy('Status', 'critical').get('length');
  }.property('failingChecks'),

  //
  // The number of passing checks within the service.
  //
  passingChecks: function() {
    return this.get('Checks').filterBy('Status', 'passing').get('length');
  }.property('passingChecks'),

  //
  // The formatted message returned for the user which represents the
  // number of checks failing or passing. Returns `1 passing` or `2 failing`
  //
  checkMessage: function() {
    if (this.get('hasFailingChecks') === false) {
      return this.get('passingChecks') + ' passing';
    } else {
      return this.get('failingChecks') + ' failing';
    }
  }.property('checkMessage'),

  //
  // Boolean of whether or not there are failing checks in the service.
  // This is used to set color backgrounds and so on.
  //
  hasFailingChecks: function() {
    return (this.get('failingChecks') > 0);
  }.property('hasFailingChecks')
});

//
// A Consul Node
//
App.Node = Ember.Object.extend({
  //
  // The number of failing checks within the service.
  //
  failingChecks: function() {
    return this.get('Checks').filterBy('Status', 'critical').get('length');
  }.property('failingChecks'),

  //
  // The number of passing checks within the service.
  //
  passingChecks: function() {
    return this.get('Checks').filterBy('Status', 'passing').get('length');
  }.property('passingChecks'),

  //
  // The formatted message returned for the user which represents the
  // number of checks failing or passing. Returns `1 passing` or `2 failing`
  //
  checkMessage: function() {
    if (this.get('hasFailingChecks') === false) {
      return this.get('passingChecks') + ' passing';
    } else {
      return this.get('failingChecks') + ' failing';
    }
  }.property('checkMessage'),

  //
  // Boolean of whether or not there are failing checks in the service.
  // This is used to set color backgrounds and so on.
  //
  hasFailingChecks: function() {
    return (this.get('failingChecks') > 0);
  }.property('hasFailingChecks')
});


//
// A key/value object
//
App.Key = Ember.Object.extend({
  isFolder: function() {
    return (this.get('key').slice(-1) == "/")
  }.property('isFolder'),

  keyParts: function() {
    var key = this.get('key');

    if (key.slice(-1) == "/") {
      key = key.substring(0, key.length - 1);
    }
    return key.split('/');
  }.property('keyParts'),

  parentKey: function() {
    var parts = this.get('keyParts');

    parts.pop();

    return parts.join("/") + "/";
  }.property('parentKey'),

  grandParentKey: function() {
    var parts = this.get('keyParts');

    parts.pop();
    parts.pop();

    return parts.join("/") + "/";
  }.property('grandParentKey')
});
