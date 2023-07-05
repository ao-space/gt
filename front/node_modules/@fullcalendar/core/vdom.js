import { __extends } from 'tslib';
import * as preact from 'preact';
import * as preactCompat from 'preact/compat';

var globalObj = typeof globalThis !== 'undefined' ? globalThis : window; // // TODO: streamline when killing IE11 support
if (globalObj.FullCalendarVDom) {
    console.warn('FullCalendar VDOM already loaded');
}
else {
    globalObj.FullCalendarVDom = {
        Component: preact.Component,
        createElement: preact.createElement,
        render: preact.render,
        createRef: preact.createRef,
        Fragment: preact.Fragment,
        createContext: createContext,
        createPortal: preactCompat.createPortal,
        flushSync: flushSync,
        unmountComponentAtNode: unmountComponentAtNode,
    };
}
// HACKS...
// TODO: lock version
// TODO: link gh issues
function flushSync(runBeforeFlush) {
    runBeforeFlush();
    var oldDebounceRendering = preact.options.debounceRendering; // orig
    var callbackQ = [];
    function execCallbackSync(callback) {
        callbackQ.push(callback);
    }
    preact.options.debounceRendering = execCallbackSync;
    preact.render(preact.createElement(FakeComponent, {}), document.createElement('div'));
    while (callbackQ.length) {
        callbackQ.shift()();
    }
    preact.options.debounceRendering = oldDebounceRendering;
}
var FakeComponent = /** @class */ (function (_super) {
    __extends(FakeComponent, _super);
    function FakeComponent() {
        return _super !== null && _super.apply(this, arguments) || this;
    }
    FakeComponent.prototype.render = function () { return preact.createElement('div', {}); };
    FakeComponent.prototype.componentDidMount = function () { this.setState({}); };
    return FakeComponent;
}(preact.Component));
function createContext(defaultValue) {
    var ContextType = preact.createContext(defaultValue);
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
    preact.render(null, node);
}
