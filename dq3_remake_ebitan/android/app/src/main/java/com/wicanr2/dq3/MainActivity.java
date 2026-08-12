package com.wicanr2.dq3;

import android.app.Activity;
import android.os.Build;
import android.os.Bundle;
import android.view.View;
import android.view.ViewGroup;
import android.view.Window;
import android.view.WindowInsets;
import android.view.WindowInsetsController;
import android.view.WindowManager;
import go.Seq;
import com.wicanr2.dq3.mobile.EbitenView; // ebitenmobile bind 產生(Go package = mobile)

public class MainActivity extends Activity {
    private EbitenView ebitenView;

    @Override
    protected void onCreate(Bundle savedInstanceState) {
        super.onCreate(savedInstanceState);
        Seq.setContext(getApplicationContext());
        ebitenView = new EbitenView(this);
        ebitenView.setLayoutParams(new ViewGroup.LayoutParams(
                ViewGroup.LayoutParams.MATCH_PARENT, ViewGroup.LayoutParams.MATCH_PARENT));
        setContentView(ebitenView);
        // 必須先掛上 content view，PhoneWindow 才會建立可供 WindowInsets 使用的 DecorView。
        applyGameChrome();
    }

    @Override protected void onPause()  {
        super.onPause();
        if (ebitenView != null) {
            ebitenView.suspendGame();
        }
    }

    @Override protected void onResume() {
        super.onResume();
        if (ebitenView != null) {
            ebitenView.resumeGame();
        }
    }

    /**
     * Android 外殼只處理平台 chrome；640x350 遊戲畫布與觸控控制仍由 Ebitengine 管理。
     * 沉浸式設定在重新取得焦點後重套，避免系統手勢暫時叫出 system bars 後留下黑邊。
     */
    private void applyGameChrome() {
        Window window = getWindow();
        window.addFlags(WindowManager.LayoutParams.FLAG_KEEP_SCREEN_ON);
        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.R) {
            window.setDecorFitsSystemWindows(false);
            WindowInsetsController controller = window.getInsetsController();
            if (controller != null) {
                controller.hide(WindowInsets.Type.systemBars());
                controller.setSystemBarsBehavior(
                        WindowInsetsController.BEHAVIOR_SHOW_TRANSIENT_BARS_BY_SWIPE);
            }
            return;
        }
        window.getDecorView().setSystemUiVisibility(
                View.SYSTEM_UI_FLAG_FULLSCREEN
                        | View.SYSTEM_UI_FLAG_HIDE_NAVIGATION
                        | View.SYSTEM_UI_FLAG_IMMERSIVE_STICKY
                        | View.SYSTEM_UI_FLAG_LAYOUT_FULLSCREEN
                        | View.SYSTEM_UI_FLAG_LAYOUT_HIDE_NAVIGATION
                        | View.SYSTEM_UI_FLAG_LAYOUT_STABLE);
    }

    @Override
    public void onWindowFocusChanged(boolean hasFocus) {
        super.onWindowFocusChanged(hasFocus);
        if (hasFocus) {
            applyGameChrome();
        }
    }
}
