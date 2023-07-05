'use strict';

Object.defineProperty(exports, '__esModule', { value: true });

/// <reference types="@fullcalendar/core-preact" />
if (typeof FullCalendarVDom === 'undefined') {
    throw new Error('Please import the top-level fullcalendar lib before attempting to import a plugin.');
}
var Component = FullCalendarVDom.Component;
var createElement = FullCalendarVDom.createElement;
var render = FullCalendarVDom.render;
var createRef = FullCalendarVDom.createRef;
var Fragment = FullCalendarVDom.Fragment;
var createContext = FullCalendarVDom.createContext;
var createPortal = FullCalendarVDom.createPortal;
var flushSync = FullCalendarVDom.flushSync;
var unmountComponentAtNode = FullCalendarVDom.unmountComponentAtNode;
/* eslint-enable */

exports.Component = Component;
exports.Fragment = Fragment;
exports.createContext = createContext;
exports.createElement = createElement;
exports.createPortal = createPortal;
exports.createRef = createRef;
exports.flushSync = flushSync;
exports.render = render;
exports.unmountComponentAtNode = unmountComponentAtNode;
