import {Component, OnInit} from '@angular/core';
import {Location} from "@angular/common";
import {Feature} from "../services/feature";
import {FeatureService} from "../services/feature.service";
import {Router} from "@angular/router";

@Component({
  selector: 'app-new-feature',
  templateUrl: './new-feature.component.html',
})
export class NewFeatureComponent implements OnInit {
  feature: Feature = {
    id: null,
    technicalName: '',
    displayName: null,
    description: null,
    expiresOn: null,
    inverted: false,
    createdAt: 0,
    updatedAt: 0,
    customerIds: null,
  };

  loading = false;

  constructor(
    private featureService: FeatureService,
    private location: Location,
    private router: Router
  ) {
  }

  ngOnInit(): void {
  }

  goBack(): void {
    this.location.back();
  }

  invert(): void {
    this.feature.inverted = !this.feature.inverted;
  }

  getExpiredOnDatetime(): string | null {
    if (this.feature.expiresOn === null) {
      return null
    }
    const utc = new Date(this.feature.expiresOn);
    return new Date(utc.getTime() - utc.getTimezoneOffset()*60*1000)
      .toISOString()
      .slice(0, -1);
  }

  setExpiredOnDatetime(e: any): void {
    this.feature.expiresOn = new Date(e.target.value).valueOf();
  }

  saveFeature(): void {
    this.loading = true;
    const that = this;
    this.featureService.saveFeature(this.feature)
      .subscribe({
        complete() {
          that.router.navigate(['/features']);
        },

        error(e) {
          console.log(e);
          that.loading = false;
        }
      });
  }
}
