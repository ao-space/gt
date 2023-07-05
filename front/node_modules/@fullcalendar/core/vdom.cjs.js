'use strict';

var tslib = require('tslib');
var preact = require('preact');
var preactCompat = require('preact/compat');

function _interopNamespace(e) {
    if (e && e.__esModule) return e;
    var n = Object.create(null);
    if (e) {
        Object.keys(e).forEach(function (k) {
            if (k !== 'default') {
                var d = Object.getOwnPropertyDescriptor(e, k);
                Object.defineProperty(n, k, d.get ? d : {
                    enumerable: true,
                    get: function () {
                        return e[k];
                    }
                });
            }
        });
    }
    n['default'] = e;
    return Object.freeze(n);
}

var preact__namespace = /*#__PURE__*/_interopNamespace(preact);
var preactCompat__namespace = /*#__PURE__*/_interopNamespace(preactCompat);

var globalObj = typeof globalThis !== 'undefined' ? globalThis : window; // // TODO: streamline when killing IE11 support
if (globalObj.FullCalendarVDom) {
    console.warn('FullCalendar VDOM already loaded');
}
else {
    globalObj.FullCalendarVDom = {
        Component: preact__namespace.Component,
        createElement: preact__namespace.createElement,
        render: preact__namespace.render,
        createRef: preact__namespace.createRef,
        Fragment: preact__namespace.Fragment,
        createContext: createContext,
        createPortal: preactCompat__namespace.createPortal,
        flushSync: flushSync,
        unmountComponentAtNode: unmountComponentAtNode,
    };
}
// HACKS...
// TODO: lock version
// TODO: link gh issues
function flushSync(runBeforeFlush) {
    runBeforeFlush();
    var oldDebounceRendering = preact__namespace.options.debounceRendering; // orig
    var callbackQ = [];
    function execCallbackSync(callback) {
        callbackQ.push(callback);
    }
    preact__namespace.options.debounceRendering = execCallbackSync;
    preact__namespace.render(preact__namespace.createElement(FakeComponent, {}), document.createElement('div'));
    while (callbackQ.length) {
        callbackQ.shift()();
    }
    preact__namespace.options.debounceRendering = oldDebounceRendering;
}
var FakeComponent = /** @class */ (function (_super) {
    tslib.__extends(FakeComponent, _super);
    function FakeComponent() {
        return _super !== null && _super.apply(this, arguments) || this;
    }
    FakeComponent.prototype.render = function () { return preact__namespace.createElement('div', {}); };
    FakeComponent.prototype.componentDidMount = function () { this.setState({}); };
    return FakeComponent;
}(preact__namespace.Component));
function createContext(defaultValue) {
    var ContextType = preact__namespace.createContext(defaultValue);
    var origProvider = ContextType.Provider;
    ContextType.Provider = function () {
        var _this = this;
        var isNew = !this.getChildContext;
        var children = origProvider.apply(this, arguments); // eslint-disable-line prefer-rest-params
        if (isNew) {
            var subs_1 = [];
            this.shouldComponentUpdate = function (_props) {
                if (_this.props.value !== _props.value) {
                    subs_1.forEach(function (c) {
                        c.context = _props.value;
                        c.forceUpdate();
                    });
                }
            };
            this.sub = function (c) {
                subs_1.push(c);
                var old = c.componentWillUnmount;
                c.componentWillUnmount = function () {
                    subs_1.splice(subs_1.indexOf(c), 1);
                    old && old.call(c);
                };
            };
        }
        return children;
    };
    return ContextType;
}
function unmountComponentAtNode(node) {
    preact__namespace.render(null, node);
}
