import NProgress from "nprogress";
import "nprogress/nprogress.css";

NProgress.configure({
  easing: "ease", // animation style
  speed: 500, // the increase speed of the progress bar
  showSpinner: true, // whether to display the loading icon
  trickleSpeed: 200, // auto-increment interval
  minimum: 0.3 // 初始化时的最小百分比
});

export default NProgress;
