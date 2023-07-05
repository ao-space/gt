import { createApp, h } from 'vue';
import { createPlugin } from '@fullcalendar/core';
/*
wrap it in an object with a `vue` key, which the custom content-type handler system will look for
*/
export function wrapVDomGenerator(vDomGenerator) {
    return function (props) {
        return { vue: vDomGenerator(props) };
    };
}
export function createVueContentTypePlugin(appContext) {
    return createPlugin({
        contentTypeHandlers: {
            vue: () => buildVDomHandler(appContext), // looks for the `vue` key
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
            app = initApp(vDomContent, appContext);
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
    return createApp({
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
                return h('span', {}, content);
            }
        }
    });
}
//# sourceMappingURL=custom-content-type.js.map