'use strict';

Object.defineProperty(exports, '__esModule', { value: true });

var enAu = {
  code: "en-au",
  week: {
    dow: 1,
    doy: 4
  },
  buttonHints: {
    prev: "Previous $0",
    next: "Next $0",
    today: "This $0"
  },
  viewHint: "$0 view",
  navLinkHint: "Go to $0",
  moreLinkHint: function(eventCnt) {
    return "Show ".concat(eventCnt, " more event").concat(eventCnt === 1 ? "" : "s");
  }
};

exports.default = enAu;
