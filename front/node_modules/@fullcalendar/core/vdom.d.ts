import * as preact from 'preact';
import * as preactCompat from 'preact/compat';
declare global {
    namespace FullCalendarVDom {
        export import Ref = preact.Ref;
        export import RefObject = preact.RefObject;
        export import ComponentType = preact.ComponentType;
        type VNode = preact.VNode<any>;
        export import Context = preact.Context;
        export import Component = preact.Component;
        export import ComponentChild = preact.ComponentChild;
        export import ComponentChildren = preact.ComponentChildren;
        export import createElement = preact.createElement;
        export import render = preact.render;
        export import createRef = preact.createRef;
        export import Fragment = preact.Fragment;
        export import createContext = preact.createContext;
        export import createPortal = preactCompat.createPortal;
        type VUIEvent = UIEvent;
        function flushSync(callback: () => void): void;
        function unmountComponentAtNode(node: HTMLElement): void;
    }
}
