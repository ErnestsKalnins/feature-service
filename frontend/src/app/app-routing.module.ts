import {NgModule} from '@angular/core';
import {RouterModule, Routes} from '@angular/router';
import {NewFeatureComponent} from "./new-feature/new-feature.component";
import {FeatureListComponent} from "./feature-list/feature-list.component";

const routes: Routes = [
  {path: '', component: FeatureListComponent},
  {path: 'new', component: NewFeatureComponent},
];

@NgModule({
  imports: [RouterModule.forRoot(routes)],
  exports: [RouterModule]
})
export class AppRoutingModule {
}
