import { defineComponent, h } from 'vue';
import { Calendar } from '@fullcalendar/core';
import { OPTION_IS_COMPLEX } from './options';
import { shallowCopy, mapHash } from './utils';
import { wrapVDomGenerator, createVueContentTypePlugin } from './custom-content-type';
const FullCalendar = defineComponent({
    props: {
        options: Object
    },
    data: initData,
    render() {
        return h('div', {
            // when renderId is changed, Vue will trigger a real-DOM async rerender, calling beforeUpdate/updated
            attrs: { 'data-fc-render-id': this.renderId }
        });
    },
    mounted() {
        // store internal data (slotOptions, calendar)
        // https://github.com/vuejs/vue/issues/1988#issuecomment-163013818
        this.slotOptions = mapHash(this.$slots, wrapVDomGenerator); // needed for buildOptions
        let calendarOptions = this.buildOptions(this.options, this.$.appContext);
        let calendar = new Calendar(this.$el, calendarOptions);
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
export default FullCalendar;
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
            createVueContentTypePlugin(appContext)
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
//# sourceMappingURL=FullCalendar.js.map