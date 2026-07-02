package com.wicanr2.dq3;

import android.app.Activity;
import android.os.Bundle;
import android.view.ViewGroup;
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
    }

    @Override protected void onPause()  { super.onPause();  ebitenView.suspendGame(); }
    @Override protected void onResume() { super.onResume(); ebitenView.resumeGame(); }
}
