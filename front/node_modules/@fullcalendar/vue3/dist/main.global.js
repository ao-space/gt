var FullCalendarVue = (function (exports, vue, core) {
    'use strict';

    const OPTION_IS_COMPLEX = {
        headerToolbar: true,
        footerToolbar: true,
        events: true,
        eventSources: true,
        resources: true
    };

    // TODO: add types!
    /*
    works with objects and arrays
    */
    function shallowCopy(val) {
        if (typeof val === 'object') {
            if (Array.isArray(val)) {
                val = Array.prototype.slice.call(val);
            }
            else if (val) { // non-null
                val = { ...val };
            }
        }
        return val;
    }
    function mapHash(input, func) {
        const output = {};
        for (const key in input) {
            if (input.hasOwnProperty(key)) {
                output[key] = func(input[key], key);
            }
        }
        return output;
    }

    /*
    wrap it in an object with a `vue` key, which the custom content-type handler system will look for
    */
    function wrapVDomGenerator(vDomGenerator) {
        return function (props) {
            return { vue: vDomGenerator(props) };
        };
    }
    function createVueContentTypePlugin(appContext) {
        return core.createPlugin({
            contentTypeHandlers: {
                vue: () => buildVDomHandler(), // looks for the `vue` key
            }
        });
    }
    function buildVDomHandler(appContext) {
        let currentEl;
        let app;
        let componentInstance;
        function render(el, vDomContent) {
            if (currentEl !== el) {
                if (currentEl && app) { // if changing elements, recreate the vue
                    app.unmount();
                }
                currentEl = el;
            }
            if (!app) {
                app = initApp(vDomContent);
                // vue's mount method *replaces* the given element. create an artificial inner el
                let innerEl = document.createElement('span');
                el.appendChild(innerEl);
                componentInstance = app.mount(innerEl);
            }
            else {
                componentInstance.content = vDomContent;
            }
        }
        function destroy() {
            if (app) { // needed?
                app.unmount();
            }
        }
        return { render, destroy };
    }
    function initApp(initialContent, appContext) {
        // TODO: do something with appContext
        return vue.createApp({
            data() {
                return {
                    content: initialContent,
                };
            },
            render() {
                let { content } = this;
                // the slot result can be an array, but the returned value of a vue component's
                // render method must be a single node.
                if (content.length === 1) {
                    return content[0];
                }
                else {
                    return vue.h('span', {}, content);
                }
            }
        });
    }

    const FullCalendar = vue.defineComponent({
        props: {
            options: Object
        },
        data: initData,
        render() {
            return vue.h('div', {
                // when renderId is changed, Vue will trigger a real-DOM async rerender, calling beforeUpdate/updated
                attrs: { 'data-fc-render-id': this.renderId }
            });
        },
        mounted() {
            // store internal data (slotOptions, calendar)
            // https://github.com/vuejs/vue/issues/1988#issuecomment-163013818
            this.slotOptions = mapHash(this.$slots, wrapVDomGenerator); // needed for buildOptions
            let calendarOptions = this.buildOptions(this.options, this.$.appContext);
            let calendar = new core.Calendar(this.$el, calendarOptions);
            this.calendar = calendar;
            calendar.render();
        },
        methods: {
            getApi,
            buildOptions,
        },
        beforeUpdate() {
            this.getApi().resumeRendering(); // the watcher handlers paused it
        },
        beforeUnmount() {
            this.getApi().destroy();
        },
        watch: buildWatchers()
    });
    function initData() {
        return {
            renderId: 0
        };
    }
    function buildOptions(suppliedOptions, appContext) {
        suppliedOptions = suppliedOptions || {};
        return {
            ...this.slotOptions,
            ...suppliedOptions,
            plugins: (suppliedOptions.plugins || []).concat([
                createVueContentTypePlugin()
            ])
        };
    }
    function getApi() {
        return this.calendar;
    }
    function buildWatchers() {
        let watchers = {
            // watches changes of ALL options and their nested objects,
            // but this is only a means to be notified of top-level non-complex options changes.
            options: {
                deep: true,
                handler(options) {
                    let calendar = this.getApi();
                    calendar.pauseRendering();
                    let calendarOptions = this.buildOptions(options, this.$.appContext);
                    calendar.resetOptions(calendarOptions);
                    this.renderId++; // will queue a rerender
                }
            }
        };
        for (let complexOptionName in OPTION_IS_COMPLEX) {
            // handlers called when nested objects change
            watchers[`options.${complexOptionName}`] = {
                deep: true,
                handler(val) {
                    // unfortunately the handler is called with undefined if new props were set, but the complex one wasn't ever set
                    if (val !== undefined) {
                        let calendar = this.getApi();
                        calendar.pauseRendering();
                        calendar.resetOptions({
                            // the only reason we shallow-copy is to trick FC into knowing there's a nested change.
                            // TODO: future versions of FC will more gracefully handle event option-changes that are same-reference.
                            [complexOptionName]: shallowCopy(val)
                        }, true);
                        this.renderId++; // will queue a rerender
                    }
                }
            };
        }
        return watchers;
    }

    exports.default = FullCalendar;
    Object.keys(core).forEach(function (k) {
        if (k !== 'default' && !exports.hasOwnProperty(k)) Object.defineProperty(exports, k, {
            enumerable: true,
            get: function () {
                return core[k];
            }
        });
    });

    Object.defineProperty(exports, '__esModule', { value: true });

    return exports;

}({}, Vue, FullCalendar));
